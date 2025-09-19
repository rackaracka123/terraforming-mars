# Scripts

TypeScript ESModule scripts for various development tasks.

## Image Converter

Converts PNG images to WebP format with 4:3 aspect ratio and 720p resolution.

### Usage

```bash
npm run convert <input-folder> <output-folder>
```

### Example

```bash
npm run convert ./my-png-images-folder ./converted-images
```

### Features

- Converts all PNG files in a directory to WebP format
- Crops images to 4:3 aspect ratio
- Scales to 960x720 resolution (720p)
- Uses ui.toast.com API for image processing
- Outputs with 85% quality for optimal file size

### Requirements

- Node.js 18+
- Internet connection for ui.toast.com API

### Installation

```bash
npm install
```