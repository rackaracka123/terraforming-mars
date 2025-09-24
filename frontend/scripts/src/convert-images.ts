#!/usr/bin/env tsx

import fs from 'fs';
import path from 'path';
import sharp from 'sharp';
import { glob } from 'glob';

interface ConversionOptions {
  aspectRatio: { width: number; height: number };
  resolution: { width: number; height: number };
  quality: number;
}

async function convertImage(imagePath: string, outputDir: string, options: ConversionOptions): Promise<boolean> {
  try {
    console.log(`Processing: ${imagePath}`);

    // If image already converted in output directory, skip it
    const fileName = path.basename(imagePath, path.extname(imagePath));
    const outputPath = path.join(outputDir, `${fileName}.webp`);
    if (fs.existsSync(outputPath)) {
      console.log(`⚠️  Skipping (already converted): ${outputPath}`);
      return true;
    }

    // Get image metadata first to calculate crop dimensions
    const image = sharp(imagePath);
    const metadata = await image.metadata();

    if (!metadata.width || !metadata.height) {
      console.error(`❌ Failed to get metadata for ${imagePath}`);
      return false;
    }

    // Calculate the crop dimensions to achieve target aspect ratio
    const targetAspectRatio = options.aspectRatio.width / options.aspectRatio.height;
    const currentAspectRatio = metadata.width / metadata.height;

    let cropWidth = metadata.width;
    let cropHeight = metadata.height;

    if (currentAspectRatio > targetAspectRatio) {
      // Image is wider than target, crop width
      cropWidth = Math.round(metadata.height * targetAspectRatio);
    } else if (currentAspectRatio < targetAspectRatio) {
      // Image is taller than target, crop height
      cropHeight = Math.round(metadata.width / targetAspectRatio);
    }

    // Calculate crop offset to center the crop
    const cropLeft = Math.round((metadata.width - cropWidth) / 2);
    const cropTop = Math.round((metadata.height - cropHeight) / 2);

    // Process the image: crop to aspect ratio, resize to target resolution, convert to WebP
    await sharp(imagePath)
      .extract({
        left: cropLeft,
        top: cropTop,
        width: cropWidth,
        height: cropHeight
      })
      .resize(options.resolution.width, options.resolution.height, {
        fit: 'fill'
      })
      .webp({
        quality: options.quality
      })
      .toFile(outputPath);

    console.log(`✅ Converted: ${outputPath}`);
    return true;

  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : 'Unknown error';
    console.error(`❌ Error converting ${imagePath}:`, errorMessage);
    return false;
  }
}

async function processFolder(inputDir: string, outputDir: string, options: ConversionOptions): Promise<void> {
  try {
    // Ensure output directory exists
    if (!fs.existsSync(outputDir)) {
      fs.mkdirSync(outputDir, { recursive: true });
      console.log(`📁 Created output directory: ${outputDir}`);
    }

    // Find all PNG files in the input directory
    const pngFiles = await glob(path.join(inputDir, '**/*.png'));

    if (pngFiles.length === 0) {
      console.log('❌ No PNG files found in the input directory');
      return;
    }

    console.log(`🖼️  Found ${pngFiles.length} PNG files to convert`);
    console.log(`📥 Input directory: ${inputDir}`);
    console.log(`📤 Output directory: ${outputDir}`);
    console.log(`🎯 Target format: WebP, ${options.resolution.width}x${options.resolution.height} (${options.aspectRatio.width}:${options.aspectRatio.height} aspect ratio)\n`);

    let successCount = 0;
    let failCount = 0;

    // Process images one by one to avoid overwhelming the API
    for (const pngFile of pngFiles) {
      const success = await convertImage(pngFile, outputDir, options);
      if (success) {
        successCount++;
      } else {
        failCount++;
      }

      // Add a small delay between API calls to be respectful
      await new Promise(resolve => setTimeout(resolve, 500));
    }

    console.log(`\n🎉 Conversion complete!`);
    console.log(`✅ Successfully converted: ${successCount} images`);
    console.log(`❌ Failed to convert: ${failCount} images`);

  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : 'Unknown error';
    console.error('❌ Error processing folder:', errorMessage);
    process.exit(1);
  }
}

function parseAspectRatio(aspectStr: string): { width: number; height: number } {
  const parts = aspectStr.split(':');
  if (parts.length !== 2) {
    throw new Error(`Invalid aspect ratio format: ${aspectStr}. Use format like "4:3" or "16:9"`);
  }

  const width = parseInt(parts[0], 10);
  const height = parseInt(parts[1], 10);

  if (isNaN(width) || isNaN(height) || width <= 0 || height <= 0) {
    throw new Error(`Invalid aspect ratio values: ${aspectStr}. Values must be positive integers`);
  }

  return { width, height };
}

function parseResolution(resolutionStr: string): { width: number; height: number } {
  const parts = resolutionStr.split('x');
  if (parts.length !== 2) {
    throw new Error(`Invalid resolution format: ${resolutionStr}. Use format like "960x720" or "1920x1080"`);
  }

  const width = parseInt(parts[0], 10);
  const height = parseInt(parts[1], 10);

  if (isNaN(width) || isNaN(height) || width <= 0 || height <= 0) {
    throw new Error(`Invalid resolution values: ${resolutionStr}. Values must be positive integers`);
  }

  return { width, height };
}

function showUsage(): void {
  console.log(`
🖼️  Image Converter - PNG to WebP with Configurable Options

Usage: npm run convert <input-folder> <output-folder> [options]

Examples:
  npm run convert ./my-png-images-folder ./converted-images
  npm run convert ./images ./output --aspect-ratio 16:9 --resolution 1920x1080
  npm run convert ./images ./output --aspect-ratio 1:1 --resolution 512x512 --quality 95

Description:
  Converts all PNG images in the input folder to WebP format with configurable
  aspect ratio, resolution, and quality settings.

Arguments:
  input-folder   Path to folder containing PNG images
  output-folder  Path where converted WebP images will be saved

Options:
  --aspect-ratio <ratio>    Target aspect ratio in "X:Y" format (default: "4:3")
  --resolution <res>        Target resolution in "WIDTHxHEIGHT" format (default: "960x720")
  --quality <number>        WebP quality from 1-100 (default: 85)

Features:
  - Smart center cropping to achieve target aspect ratio
  - Configurable output resolution
  - WebP compression with adjustable quality
  - Skip already converted files automatically
  - Uses Sharp library for fast local image processing
`);
}

// Main execution
async function main(): Promise<void> {
  const args = process.argv.slice(2);

  if (args.length < 2) {
    showUsage();
    process.exit(1);
  }

  // Parse positional arguments
  const inputDir = args[0];
  const outputDir = args[1];

  // Default options
  const defaultOptions: ConversionOptions = {
    aspectRatio: { width: 4, height: 3 },
    resolution: { width: 960, height: 720 },
    quality: 85
  };

  // Parse optional arguments
  let options = { ...defaultOptions };

  for (let i = 2; i < args.length; i++) {
    const arg = args[i];

    if (arg === '--aspect-ratio' || arg === '-a') {
      if (i + 1 >= args.length) {
        console.error('❌ --aspect-ratio requires a value (e.g., "4:3", "16:9")');
        process.exit(1);
      }
      try {
        options.aspectRatio = parseAspectRatio(args[i + 1]);
        i++; // Skip the next argument since we consumed it
      } catch (error) {
        console.error(`❌ ${error instanceof Error ? error.message : 'Invalid aspect ratio'}`);
        process.exit(1);
      }
    } else if (arg === '--resolution' || arg === '-r') {
      if (i + 1 >= args.length) {
        console.error('❌ --resolution requires a value (e.g., "960x720", "1920x1080")');
        process.exit(1);
      }
      try {
        options.resolution = parseResolution(args[i + 1]);
        i++; // Skip the next argument since we consumed it
      } catch (error) {
        console.error(`❌ ${error instanceof Error ? error.message : 'Invalid resolution'}`);
        process.exit(1);
      }
    } else if (arg === '--quality' || arg === '-q') {
      if (i + 1 >= args.length) {
        console.error('❌ --quality requires a value (1-100)');
        process.exit(1);
      }
      const quality = parseInt(args[i + 1], 10);
      if (isNaN(quality) || quality < 1 || quality > 100) {
        console.error('❌ Quality must be a number between 1 and 100');
        process.exit(1);
      }
      options.quality = quality;
      i++; // Skip the next argument since we consumed it
    } else if (arg === '--help' || arg === '-h') {
      showUsage();
      process.exit(0);
    } else {
      console.error(`❌ Unknown option: ${arg}`);
      showUsage();
      process.exit(1);
    }
  }

  // Validate input directory
  if (!fs.existsSync(inputDir)) {
    console.error(`❌ Input directory does not exist: ${inputDir}`);
    process.exit(1);
  }

  if (!fs.statSync(inputDir).isDirectory()) {
    console.error(`❌ Input path is not a directory: ${inputDir}`);
    process.exit(1);
  }

  await processFolder(path.resolve(inputDir), path.resolve(outputDir), options);
}

// Handle script execution
if (import.meta.url === `file://${process.argv[1]}`) {
  main().catch((error: unknown) => {
    const errorMessage = error instanceof Error ? error.message : 'Unknown error';
    console.error('❌ Script failed:', errorMessage);
    process.exit(1);
  });
}