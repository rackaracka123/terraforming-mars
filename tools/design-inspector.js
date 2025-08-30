const puppeteer = require('puppeteer');
const fs = require('fs').promises;
const path = require('path');
const pixelmatch = require('pixelmatch');
const { PNG } = require('pngjs');

class DesignInspector {
  constructor() {
    this.browser = null;
    this.page = null;
    this.screenshotDir = path.join(process.cwd(), 'design-screenshots');
    this.referenceDir = path.join(this.screenshotDir, 'reference');
    this.currentDir = path.join(this.screenshotDir, 'current');
    this.diffDir = path.join(this.screenshotDir, 'diff');
  }

  async init() {
    // Create directories
    await fs.mkdir(this.screenshotDir, { recursive: true });
    await fs.mkdir(this.referenceDir, { recursive: true });
    await fs.mkdir(this.currentDir, { recursive: true });
    await fs.mkdir(this.diffDir, { recursive: true });

    // Launch browser
    this.browser = await puppeteer.launch({
      headless: false,
      defaultViewport: { width: 1920, height: 1080 },
      args: ['--no-sandbox', '--disable-setuid-sandbox']
    });
    
    this.page = await this.browser.newPage();
    await this.page.setViewport({ width: 1920, height: 1080 });
  }

  async captureGameInterface(name = 'game-interface') {
    if (!this.page) throw new Error('Browser not initialized. Call init() first.');
    
    // Navigate to the game
    await this.page.goto('http://localhost:3000', { 
      waitUntil: 'networkidle2',
      timeout: 30000 
    });

    // Wait for the page to be ready - try canvas first, then fallback to basic page load
    try {
      await this.page.waitForSelector('canvas', { timeout: 10000 });
      console.log('3D canvas detected - waiting for rendering...');
      await new Promise(resolve => setTimeout(resolve, 3000));
    } catch (e) {
      console.log('No canvas found - capturing page as-is...');
      await new Promise(resolve => setTimeout(resolve, 2000));
    }

    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const filename = `${name}-${timestamp}.png`;
    const filepath = path.join(this.currentDir, filename);
    
    await this.page.screenshot({
      path: filepath,
      fullPage: true
    });

    console.log(`Screenshot saved: ${filepath}`);
    return filepath;
  }

  async captureComponent(selector, name) {
    if (!this.page) throw new Error('Browser not initialized. Call init() first.');
    
    const element = await this.page.$(selector);
    if (!element) {
      throw new Error(`Element not found: ${selector}`);
    }

    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const filename = `${name}-${timestamp}.png`;
    const filepath = path.join(this.currentDir, filename);
    
    await element.screenshot({ path: filepath });
    
    console.log(`Component screenshot saved: ${filepath}`);
    return filepath;
  }

  async setReferenceImage(currentPath, referenceName) {
    const referenceFile = path.join(this.referenceDir, `${referenceName}.png`);
    await fs.copyFile(currentPath, referenceFile);
    console.log(`Reference image set: ${referenceFile}`);
    return referenceFile;
  }

  async compareWithReference(currentPath, referenceName) {
    const referenceFile = path.join(this.referenceDir, `${referenceName}.png`);
    
    try {
      await fs.access(referenceFile);
    } catch {
      console.log(`No reference image found for ${referenceName}. Use setReferenceImage() first.`);
      return null;
    }

    const img1 = PNG.sync.read(await fs.readFile(referenceFile));
    const img2 = PNG.sync.read(await fs.readFile(currentPath));
    
    const { width, height } = img1;
    const diff = new PNG({ width, height });
    
    const numDiffPixels = pixelmatch(img1.data, img2.data, diff.data, width, height, {
      threshold: 0.1,
      includeAA: false
    });
    
    const diffPercentage = (numDiffPixels / (width * height)) * 100;
    
    const diffFilename = `diff-${referenceName}-${Date.now()}.png`;
    const diffPath = path.join(this.diffDir, diffFilename);
    await fs.writeFile(diffPath, PNG.sync.write(diff));
    
    const result = {
      reference: referenceFile,
      current: currentPath,
      diff: diffPath,
      diffPixels: numDiffPixels,
      diffPercentage: diffPercentage.toFixed(2),
      passed: diffPercentage < 1 // Less than 1% difference
    };
    
    console.log(`Comparison result: ${diffPercentage.toFixed(2)}% difference`);
    console.log(`Diff image saved: ${diffPath}`);
    
    return result;
  }

  async inspectDesign(testName, options = {}) {
    const {
      setAsReference = false,
      compareWithExisting = true,
      component = null,
      selector = null
    } = options;

    try {
      let screenshotPath;
      
      if (component && selector) {
        screenshotPath = await this.captureComponent(selector, component);
      } else {
        screenshotPath = await this.captureGameInterface(testName);
      }

      if (setAsReference) {
        await this.setReferenceImage(screenshotPath, testName);
        return { status: 'reference_set', path: screenshotPath };
      }

      if (compareWithExisting) {
        const comparison = await this.compareWithReference(screenshotPath, testName);
        return { status: 'compared', path: screenshotPath, comparison };
      }

      return { status: 'captured', path: screenshotPath };
    } catch (error) {
      console.error('Design inspection failed:', error);
      throw error;
    }
  }

  async close() {
    if (this.browser) {
      await this.browser.close();
    }
  }
}

// CLI usage
if (require.main === module) {
  const inspector = new DesignInspector();
  
  async function main() {
    const args = process.argv.slice(2);
    const command = args[0] || 'capture';
    const testName = args[1] || 'default';
    
    try {
      await inspector.init();
      
      switch (command) {
        case 'capture':
          await inspector.inspectDesign(testName);
          break;
        case 'reference':
          await inspector.inspectDesign(testName, { setAsReference: true });
          break;
        case 'compare':
          await inspector.inspectDesign(testName, { compareWithExisting: true });
          break;
        case 'component':
          const selector = args[2];
          if (!selector) throw new Error('Component selector required');
          await inspector.inspectDesign(testName, { 
            component: testName, 
            selector,
            compareWithExisting: true 
          });
          break;
        default:
          console.log('Usage: node design-inspector.js [capture|reference|compare|component] <testName> [selector]');
      }
    } finally {
      await inspector.close();
    }
  }
  
  main().catch(console.error);
}

module.exports = DesignInspector;