#!/usr/bin/env bash
# Build Go Core as shared library for mobile FFI
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
OUTPUT_DIR="${OUTPUT_DIR:-$PROJECT_ROOT/apps/frontend/native}"

# Parse arguments
ARCH=${1:-arm64}
OUTPUT=${2:-$OUTPUT_DIR}

echo "Building Go Core as shared library for mobile..."
echo "Architecture: $ARCH"
echo "Output: $OUTPUT"

# Create output directory
mkdir -p "$OUTPUT"

# Build for Android (SO file)
if [[ "$ARCH" == "arm64" ]] || [[ "$ARCH" == "all" ]]; then
  echo "Building for Android ARM64..."
  cd "$PROJECT_ROOT/packages/backend"

  GOOS=android GOARCH=arm64 CGO_ENABLED=1 \
    go build -buildmode=c-shared \
    -o "$OUTPUT/libcore_android_arm64.so" \
    ./cmd/mobile
fi

# Build for iOS (Framework)
if [[ "$ARCH" == "arm64" ]] || [[ "$ARCH" == "all" ]]; then
  echo "Building for iOS ARM64..."
  cd "$PROJECT_ROOT/packages/backend"

  GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 \
    go build -buildmode=c-archive \
    -o "$OUTPUT/libcore_ios_arm64.a" \
    ./cmd/mobile

  # Create framework header
  cat > "$OUTPUT/MemonexusCore.h" << 'EOF'
// MemoNexus Core FFI Interface
#ifndef MEMONEXUS_CORE_H
#define MEMONEXUS_CORE_H

#ifdef __cplusplus
extern "C" {
#endif

// Content operations
char* ContentCreate(const char* title, const char* content, const char* mediaType);
char* ContentList(const char* filter);
char* ContentGet(const char* id);
char* ContentUpdate(const char* id, const char* title, const char* content);
char* ContentDelete(const char* id);

// Search operations
char* SearchQuery(const char* query, const char* filters);

// Analysis operations
char* AnalyzeKeywords(const char* content);
char* GenerateSummary(const char* content);

// Utility
void FreeString(char* s);

#ifdef __cplusplus
}
#endif

#endif // MEMONEXUS_CORE_H
EOF
fi

echo "Mobile library build complete!"
echo "Output: $OUTPUT"
