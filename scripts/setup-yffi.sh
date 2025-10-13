#!/bin/bash
# scripts/setup-yffi.sh - Setup official yffi from y-crdt repository for ygo module

set -e

echo "🦀 Setting up official yffi for ygo module..."

# Create lib directory
mkdir -p lib

# Y-CRDT version to use (pinned for reproducibility)
YCRDT_VERSION="v0.24.0"

# Check if y-crdt directory exists
if [ ! -d "y-crdt" ]; then
    echo "📥 Cloning y-crdt repository (${YCRDT_VERSION})..."
    git clone --branch ${YCRDT_VERSION} --depth 1 https://github.com/y-crdt/y-crdt.git
    echo "✅ Cloned y-crdt ${YCRDT_VERSION}"
else
    echo "📁 y-crdt directory exists, verifying version..."
    cd y-crdt
    CURRENT_TAG=$(git describe --tags --exact-match 2>/dev/null || echo "unknown")
    if [ "$CURRENT_TAG" != "$YCRDT_VERSION" ]; then
        echo "⚠️  Warning: y-crdt is at $CURRENT_TAG but expected $YCRDT_VERSION"
        echo "💡 To update, run: rm -rf y-crdt && make setup-yffi"
    else
        echo "✅ y-crdt is at correct version: $YCRDT_VERSION"
    fi
    cd ..
fi

echo "🔧 Building official yffi library (static, size-optimized)..."
cd y-crdt/yffi

# Build the yffi library as static library
# First, check if Cargo.toml has staticlib in crate-type
if ! grep -q "staticlib" Cargo.toml; then
    echo "⚠️  Adding staticlib to crate-type in Cargo.toml..."
    # Backup original Cargo.toml
    cp Cargo.toml Cargo.toml.backup

    # Add staticlib to crate-type if not present
    if grep -q "crate-type" Cargo.toml; then
        # crate-type exists, add staticlib if not present
        sed -i.bak 's/crate-type = \[/crate-type = ["staticlib", /' Cargo.toml
    else
        # No crate-type, add it under [lib]
        sed -i.bak '/\[lib\]/a\
crate-type = ["staticlib", "cdylib"]' Cargo.toml
    fi
fi

# Build with size optimization using environment variables (no Cargo.toml modifications needed)
echo "📦 Building with size optimization (opt-level=z, LTO, strip)..."
CARGO_PROFILE_RELEASE_OPT_LEVEL=z \
CARGO_PROFILE_RELEASE_LTO=true \
CARGO_PROFILE_RELEASE_CODEGEN_UNITS=1 \
CARGO_PROFILE_RELEASE_STRIP=true \
CARGO_PROFILE_RELEASE_PANIC=abort \
cargo build --release

echo "📋 Copying library and header files..."

# Determine the library extension based on OS
if [[ "$OSTYPE" == "darwin"* ]]; then
    LIB_EXT="dylib"
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    LIB_EXT="so"
else
    echo "❌ Unsupported OS: $OSTYPE"
    exit 1
fi

# Copy the static library (.a file) from release build
if [ -f "../target/release/libyrs.a" ]; then
    cp ../target/release/libyrs.a ../../lib/
    echo "✅ Copied size-optimized static library: libyrs.a"
else
    echo "❌ Static library not found: ../target/release/libyrs.a"
    exit 1
fi

# Optionally copy the shared library for reference
if [ -f "../target/release/libyrs.${LIB_EXT}" ]; then
    cp ../target/release/libyrs.${LIB_EXT} ../../lib/
    echo "✅ Copied shared library: libyrs.${LIB_EXT}"
fi

# Copy the header from tests-ffi/include
cp ../tests-ffi/include/libyrs.h ../../lib/

cd ../..

echo "✅ Official yffi setup complete!"
echo "📁 Static library: lib/libyrs.a"
echo "📁 Header: lib/libyrs.h"

# Verify files exist
if [ -f "lib/libyrs.a" ] && [ -f "lib/libyrs.h" ]; then
    echo "🎉 All files copied successfully!"
    echo ""
    echo "Static linking enabled - your binaries will not depend on external .so files"
    ls -la lib/
else
    echo "❌ Setup failed - missing files"
    exit 1
fi