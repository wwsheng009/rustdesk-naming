#!/bin/sh


# host:ID服务器
# relay:中继服务器
# api:API服务器
# Key : 密钥


# 函数：检查并提示安装缺失的软件包
check_dependencies() {
    missing_deps=""
    
    # 检查 jq
    if ! command -v jq >/dev/null 2>&1; then
        missing_deps="$missing_deps jq"
    fi
    
    # 检查 base64
    if ! command -v base64 >/dev/null 2>&1; then
        missing_deps="$missing_deps base64(coreutils)"
    fi
    
    # 检查 rev
    if ! command -v rev >/dev/null 2>&1; then
        missing_deps="$missing_deps rev(coreutils)"
    fi

    # 如果有缺失的依赖，提示用户安装
    if [ -n "$missing_deps" ]; then
        echo "Error: The following dependencies are missing:$missing_deps"
        echo "Please install them using the appropriate command for your system:"
        
        # 检测系统类型
        if [ "$(uname)" = "Darwin" ]; then
            echo "For macOS:"
            echo "  brew install jq coreutils"
        elif [ -f /etc/debian_version ]; then
            echo "For Ubuntu/Debian:"
            echo "  sudo apt-get update"
            echo "  sudo apt-get install jq coreutils"
        else
            echo "For other systems, please install jq and coreutils manually."
        fi
        exit 1
    fi
}

# 函数：生成编码后的名称
gen_name() {
    key="$1"
    host="$2"
    api="$3"
    relay="$4"

    # 创建紧凑的 JSON 对象，仅包含非空值
    json=$(jq -c -n \
        --arg key "$key" \
        --arg host "$host" \
        --arg api "$api" \
        --arg relay "$relay" \
        '{host:$host,key:$key} | if $api != "" then . + {api:$api} else . end | if $relay != "" then . + {relay:$relay} else . end')
    echo "Debug: JSON: $json"

    # TODO: Add signature if required (e.g., Ed25519 signing)
    # signed_data=$(sign_json "$json" --private-key=<key>)
    # For now, use unsigned JSON
    data_to_encode="$json"

    # 转换为 Base64（URL安全，无填充）
    base64_encoded=$(echo -n "$data_to_encode" | base64 -w 0 | tr -d '=')
    echo "Debug: Base64: $base64_encoded"

    # 反转字符串
    reversed=$(echo "$base64_encoded" | rev)
    echo "Debug: Reversed: $reversed"

    echo "$reversed"
}

# 函数：解码输入字符串
decode_name() {
    encoded="$1"

    # 反转字符串
    reversed=$(echo "$encoded" | rev)
    if [ $? -ne 0 ]; then
        echo "Error: Failed to reverse input string"
        return 1
    fi
    echo "Debug: Reversed string: $reversed"

    # Base64 解码（支持无填充）
    # 添加必要的填充字符
    padding=$((4 - ${#reversed} % 4))
    if [ $padding -ne 4 ]; then
        reversed="$reversed$(printf '=%.0s' $(seq 1 $padding))"
    fi
    decoded=$(echo "$reversed" | base64 -d 2>/dev/null)
    if [ $? -ne 0 ]; then
        echo "Error: Failed to decode Base64 string"
        echo "Debug: Invalid Base64 input: $reversed"
        return 1
    fi
    echo "Debug: Decoded string: $decoded"

    # TODO: Verify signature if signed
    # verified_data=$(verify_signature "$decoded" --public-key=<key>)
    # For now, assume unsigned JSON
    json_data="$decoded"

    # 验证是否为有效 JSON
    if ! echo "$json_data" | jq . >/dev/null 2>&1; then
        echo "Error: Decoded string is not valid JSON"
        echo "Debug: Invalid JSON: $json_data"
        return 1
    fi

    # 提取并显示字段
    key=$(echo "$json_data" | jq -r '.key')
    host=$(echo "$json_data" | jq -r '.host')
    api=$(echo "$json_data" | jq -r '.api')
    relay=$(echo "$json_data" | jq -r '.relay')

    echo "Decoded CustomServer:"
    echo "  密钥KEY: $key"
    echo "  ID服务器: $host"
    echo "  API服务器: $api"
    echo "  中继服务器: $relay"
}

# 函数：交互式输入（用于编码）
prompt_for_input() {
    echo "Enter key: "
    read key
    if [ -z "$key" ]; then
        echo "Error: key is required."
        exit 1
    fi

    echo "Enter host: "
    read host
    if [ -z "$host" ]; then
        echo "Error: host is required."
        exit 1
    fi

    echo "Enter api (optional, press Enter to skip): "
    read api
    echo "Enter relay (optional, press Enter to skip): "
    read relay

    # 调用 gen_name 生成编码
    result=$(gen_name "$key" "$host" "$api" "$relay")
    if [ $? -eq 0 ]; then
        echo "rustdesk-custom_serverd-$result.exe"
    else
        echo "Error: Failed to generate name"
        exit 1
    fi
}

# 检查依赖
check_dependencies

# 主逻辑
if [ $# -ge 2 ]; then
    # 使用命令行参数进行编码
    key="$1"
    host="$2"
    api="${3-}"
    relay="${4-}"

    # 生成编码后的名称
    result=$(gen_name "$key" "$host" "$api" "$relay")
    if [ $? -eq 0 ]; then
        echo "rustdesk-custom_serverd-$result.exe"
    else
        echo "Error: Failed to generate name"
        exit 1
    fi
elif [ $# -eq 1 ]; then
    # 如果只有一个参数，尝试解码
    # 移除可能的 rustdesk-custom_serverd- 前缀和 .exe 后缀
    input=$(echo "$1" | sed 's/^rustdesk-custom_serverd-//; s/\.exe$//')
    decode_name "$input"
    if [ $? -ne 0 ]; then
        echo "Error: Failed to decode input string"
        exit 1
    fi
else
    # 没有命令行参数，进入交互模式（用于编码）
    echo "No command-line arguments provided. Entering interactive mode."
    prompt_for_input
fi