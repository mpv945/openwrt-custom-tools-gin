#!/bin/sh
# upload-artifact.sh
# Alpine + ARM64 CI 下手动上传 GitHub Actions artifact
# Usage: ./upload-artifact.sh <artifact_name> <file_or_dir>

set -e

ARTIFACT_NAME="$1"
SRC_PATH="$2"

if [ -z "$ARTIFACT_NAME" ] || [ -z "$SRC_PATH" ]; then
  echo "Usage: $0 <artifact_name> <file_or_dir>"
  exit 1
fi

if [ -z "$GITHUB_TOKEN" ]; then
  echo "GITHUB_TOKEN is required"
  exit 1
fi

# 1. 压缩文件/目录
ZIP_FILE="${ARTIFACT_NAME}.zip"
echo "Compressing ${SRC_PATH} -> ${ZIP_FILE}..."
zip -r -q "$ZIP_FILE" "$SRC_PATH"

# 2. 获取当前 workflow run ID
echo "Fetching current workflow run ID..."
RUN_ID=$(gh run list -L 1 --json databaseId -q '.[0].databaseId')
echo "Current workflow run ID: $RUN_ID"

# 3. 创建 artifact 对象
echo "Creating artifact object..."
CREATE_RESP=$(curl -s -X POST \
  -H "Authorization: Bearer $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github+json" \
  https://api.github.com/repos/${GITHUB_REPOSITORY}/actions/artifacts \
  -d '{
    "name": "'"${ARTIFACT_NAME}"'",
    "type": "container",
    "workflow_run_id": '"${RUN_ID}"'
  }'
)

UPLOAD_URL=$(echo "$CREATE_RESP" | grep -o '"upload_url": "[^"]*' | grep -o '[^"]*$')
if [ -z "$UPLOAD_URL" ]; then
  echo "Failed to get upload_url from response:"
  echo "$CREATE_RESP"
  exit 1
fi
echo "Upload URL: $UPLOAD_URL"

# 4. 上传 artifact
echo "Uploading artifact..."
curl -X PUT \
  -H "Content-Type: application/zip" \
  --data-binary @"$ZIP_FILE" \
  "$UPLOAD_URL"

# 5. Finalize artifact
echo "Finalizing artifact..."
curl -X POST \
  -H "Authorization: Bearer $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github+json" \
  "${UPLOAD_URL}/complete"

echo "Artifact '${ARTIFACT_NAME}' uploaded successfully!"