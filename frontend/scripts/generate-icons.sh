#!/bin/bash
# Generate PWA icons from source logo
# Usage: ./scripts/generate-icons.sh <source-image>
# Example: ./scripts/generate-icons.sh logo.png

SOURCE_IMAGE="${1:-logo.png}"
OUTPUT_DIR="public"

if [ ! -f "$SOURCE_IMAGE" ]; then
    echo "Error: Source image '$SOURCE_IMAGE' not found"
    echo "Usage: ./scripts/generate-icons.sh <source-image>"
    exit 1
fi

echo "Generating PWA icons from $SOURCE_IMAGE..."

# Check if we're on macOS (use sips) or have ImageMagick
if command -v sips &> /dev/null; then
    echo "Using macOS sips..."

    # Create icons using sips
    sips -z 16 16 "$SOURCE_IMAGE" --out "$OUTPUT_DIR/favicon-16x16.png" 2>/dev/null
    sips -z 32 32 "$SOURCE_IMAGE" --out "$OUTPUT_DIR/favicon-32x32.png" 2>/dev/null
    sips -z 180 180 "$SOURCE_IMAGE" --out "$OUTPUT_DIR/apple-touch-icon.png" 2>/dev/null
    sips -z 192 192 "$SOURCE_IMAGE" --out "$OUTPUT_DIR/pwa-192x192.png" 2>/dev/null
    sips -z 512 512 "$SOURCE_IMAGE" --out "$OUTPUT_DIR/pwa-512x512.png" 2>/dev/null

elif command -v convert &> /dev/null; then
    echo "Using ImageMagick..."

    # Create icons using ImageMagick
    convert "$SOURCE_IMAGE" -resize 16x16 "$OUTPUT_DIR/favicon-16x16.png"
    convert "$SOURCE_IMAGE" -resize 32x32 "$OUTPUT_DIR/favicon-32x32.png"
    convert "$SOURCE_IMAGE" -resize 180x180 "$OUTPUT_DIR/apple-touch-icon.png"
    convert "$SOURCE_IMAGE" -resize 192x192 "$OUTPUT_DIR/pwa-192x192.png"
    convert "$SOURCE_IMAGE" -resize 512x512 "$OUTPUT_DIR/pwa-512x512.png"

else
    echo "Error: Neither sips (macOS) nor ImageMagick (convert) found"
    echo "Please install ImageMagick: brew install imagemagick"
    exit 1
fi

echo "Icons generated successfully in $OUTPUT_DIR/"
ls -la "$OUTPUT_DIR"/*.png 2>/dev/null
