#!/usr/bin/env tsx

import fs from 'fs';
import path from 'path';
import sharp from 'sharp';
import { glob } from 'glob';

async function convertImage(imagePath: string, outputDir: string): Promise<boolean> {
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

    // Calculate the crop dimensions to achieve 4:3 aspect ratio
    const targetAspectRatio = 4 / 3;
    const currentAspectRatio = metadata.width / metadata.height;

    let cropWidth = metadata.width;
    let cropHeight = metadata.height;

    if (currentAspectRatio > targetAspectRatio) {
      // Image is wider than 4:3, crop width
      cropWidth = Math.round(metadata.height * targetAspectRatio);
    } else if (currentAspectRatio < targetAspectRatio) {
      // Image is taller than 4:3, crop height
      cropHeight = Math.round(metadata.width / targetAspectRatio);
    }

    // Calculate crop offset to center the crop
    const cropLeft = Math.round((metadata.width - cropWidth) / 2);
    const cropTop = Math.round((metadata.height - cropHeight) / 2);

    // Process the image: crop to 4:3, resize to 720p, convert to WebP
    await sharp(imagePath)
      .extract({
        left: cropLeft,
        top: cropTop,
        width: cropWidth,
        height: cropHeight
      })
      .resize(960, 720, {
        fit: 'fill'
      })
      .webp({
        quality: 85
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

async function processFolder(inputDir: string, outputDir: string): Promise<void> {
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
    console.log(`🎯 Target format: WebP, 960x720 (4:3 aspect ratio)\n`);

    let successCount = 0;
    let failCount = 0;

    // Process images one by one to avoid overwhelming the API
    for (const pngFile of pngFiles) {
      const success = await convertImage(pngFile, outputDir);
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

function showUsage(): void {
  console.log(`
🖼️  Image Converter - PNG to WebP (4:3, 720p)

Usage: npm run convert <input-folder> <output-folder>

Example:
  npm run convert ./my-png-images-folder ./converted-images

Description:
  Converts all PNG images in the input folder to WebP format.
  - Crops to 4:3 aspect ratio (center crop)
  - Scales to 960x720 resolution (720p)
  - Outputs as WebP with 85% quality
  - Uses Sharp library for fast local image processing

Arguments:
  input-folder   Path to folder containing PNG images
  output-folder  Path where converted WebP images will be saved
`);
}

// Main execution
async function main(): Promise<void> {
  const args = process.argv.slice(2);

  if (args.length !== 2) {
    showUsage();
    process.exit(1);
  }

  const [inputDir, outputDir] = args;

  // Validate input directory
  if (!fs.existsSync(inputDir)) {
    console.error(`❌ Input directory does not exist: ${inputDir}`);
    process.exit(1);
  }

  if (!fs.statSync(inputDir).isDirectory()) {
    console.error(`❌ Input path is not a directory: ${inputDir}`);
    process.exit(1);
  }

  await processFolder(path.resolve(inputDir), path.resolve(outputDir));
}

// Handle script execution
if (import.meta.url === `file://${process.argv[1]}`) {
  main().catch((error: unknown) => {
    const errorMessage = error instanceof Error ? error.message : 'Unknown error';
    console.error('❌ Script failed:', errorMessage);
    process.exit(1);
  });
}