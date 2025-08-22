#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const { execSync, spawn } = require('child_process');
const { performance } = require('perf_hooks');

// Configuration
const CONFIG = {
    testUrlsFile: process.argv[2] || './testurls.txt',
    outputDir: './test-output',
    goBinary: '../bin/hermes',
    comparisonReport: './test-output/comparison-report.json'
};

// Ensure output directories exist and are clean
function setupOutputDirectories() {
    console.log('🧹 Cleaning up and setting up output directories...');

    const dirs = [
        `${CONFIG.outputDir}/js/json`,
        `${CONFIG.outputDir}/js/markdown`,
        `${CONFIG.outputDir}/go/json`,
        `${CONFIG.outputDir}/go/markdown`
    ];

    // Remove all files from output directories
    try {
        if (fs.existsSync(CONFIG.outputDir)) {
            fs.rmSync(CONFIG.outputDir, { recursive: true, force: true });
        }

        // Recreate directory structure
        dirs.forEach(dir => {
            fs.mkdirSync(dir, { recursive: true });
        });

        console.log('✅ Output directories cleaned and created');
    } catch (error) {
        console.error('❌ Error setting up directories:', error.message);
        process.exit(1);
    }
}

// Install postlight parser via npm
function installPostlightParser() {
    console.log('📦 Installing @postlight/parser from npm...');

    try {
        // Create a temporary package.json if it doesn't exist
        if (!fs.existsSync('./package.json')) {
            const packageJson = {
                name: "parser-comparison-temp",
                version: "1.0.0",
                dependencies: {}
            };
            fs.writeFileSync('./package.json', JSON.stringify(packageJson, null, 2));
        }

        // Install the parser
        execSync('npm install @postlight/parser', {
            stdio: 'inherit',
            cwd: process.cwd()
        });

        console.log('✅ @postlight/parser installed');
        return true;
    } catch (error) {
        console.error('❌ Failed to install @postlight/parser:', error.message);
        return false;
    }
}

// Build Go parser
function buildGoParser() {
    console.log('🔨 Building Go parser...');

    try {
        const parserGoDir = '..'; // Now we're in parser-go/benchmark, so parent dir is parser-go
        if (!fs.existsSync(parserGoDir)) {
            throw new Error('parser-go directory not found');
        }

        // Check if Makefile exists, otherwise use go build
        if (fs.existsSync(path.join(parserGoDir, 'Makefile'))) {
            execSync('make build', {
                cwd: parserGoDir,
                stdio: 'inherit'
            });
        } else {
            execSync('go build -o bin/hermes ./cmd/parser', {
                cwd: parserGoDir,
                stdio: 'inherit'
            });
        }

        console.log('✅ Go parser built');
        return true;
    } catch (error) {
        console.error('❌ Failed to build Go parser:', error.message);
        return false;
    }
}

// Parse URL with JavaScript parser
async function parseWithJS(url, format) {
    const startTime = performance.now();

    try {
        const Parser = require('@postlight/parser');

        let result;
        if (format === 'json') {
            result = await Parser.parse(url, { contentType: 'html' });
            result = JSON.stringify(result, null, 2);
        } else {
            result = await Parser.parse(url, { contentType: 'markdown' });
            result = result.content || '';
        }

        const endTime = performance.now();
        const executionTime = Math.round(endTime - startTime);

        return {
            success: true,
            executionTime,
            output: result
        };
    } catch (error) {
        const endTime = performance.now();
        const executionTime = Math.round(endTime - startTime);

        return {
            success: false,
            executionTime,
            error: error.message,
            output: ''
        };
    }
}

// Parse URL with Go parser
function parseWithGo(url, format) {
    const startTime = performance.now();

    try {
        const output = execSync(`"${CONFIG.goBinary}" parse --format ${format} "${url}"`, {
            encoding: 'utf8',
            timeout: 30000 // 30 second timeout
        });

        const endTime = performance.now();
        const executionTime = Math.round(endTime - startTime);

        return {
            success: true,
            executionTime,
            output: output.trim()
        };
    } catch (error) {
        const endTime = performance.now();
        const executionTime = Math.round(endTime - startTime);

        return {
            success: false,
            executionTime,
            error: error.message,
            output: ''
        };
    }
}

// Sanitize filename for cross-platform compatibility
function sanitizeFilename(url) {
    return url
        .replace(/https?:\/\//, '')
        .replace(/[^a-zA-Z0-9.-]/g, '_')
        .substring(0, 100); // Limit length
}

// Format bytes for display
function formatBytes(bytes) {
    return (bytes / 1024 / 1024).toFixed(2);
}

// Test both parsers for a specific format
async function testFormat(format, urls) {
    console.log(`\n🚀 Testing ${format.toUpperCase()} format with ${urls.length} URLs`);
    console.log('='.repeat(50));

    const results = [];
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');

    // Phase 1: Run JavaScript parser for all URLs
    console.log('\n📦 JavaScript Parser - Processing all URLs...');
    const jsResults = [];
    let jsTotal = 0, jsSuccess = 0;

    for (let i = 0; i < urls.length; i++) {
        const url = urls[i];
        process.stdout.write(`  URL ${i + 1}/${urls.length}... `);
        
        const jsResult = await parseWithJS(url, format);
        jsResults.push(jsResult);
        jsTotal += jsResult.executionTime;
        
        if (jsResult.success) {
            jsSuccess++;
            console.log(`✓ ${jsResult.executionTime}ms`);
        } else {
            console.log(`✗ ${jsResult.executionTime}ms (${jsResult.error})`);
        }
    }

    // Phase 2: Run Go parser for all URLs
    console.log('\n🔧 Go Parser - Processing all URLs...');
    const goResults = [];
    let goTotal = 0, goSuccess = 0;

    for (let i = 0; i < urls.length; i++) {
        const url = urls[i];
        process.stdout.write(`  URL ${i + 1}/${urls.length}... `);
        
        const goResult = parseWithGo(url, format);
        goResults.push(goResult);
        goTotal += goResult.executionTime;
        
        if (goResult.success) {
            goSuccess++;
            console.log(`✓ ${goResult.executionTime}ms`);
        } else {
            console.log(`✗ ${goResult.executionTime}ms (${goResult.error})`);
        }
    }

    // Phase 3: Save files and compile results
    console.log('\n💾 Saving results and compiling report...');
    
    for (let i = 0; i < urls.length; i++) {
        const url = urls[i];
        const filename = sanitizeFilename(url);
        const jsRes = jsResults[i];
        const goRes = goResults[i];
        
        // Save JavaScript result
        let jsSizeBytes = 0;
        if (jsRes.success) {
            const jsFile = path.join(CONFIG.outputDir, 'js', format, `${filename}-${timestamp}-${i}.${format}`);
            fs.writeFileSync(jsFile, jsRes.output);
            jsSizeBytes = fs.statSync(jsFile).size;
        }

        // Save Go result
        let goSizeBytes = 0;
        if (goRes.success) {
            const goFile = path.join(CONFIG.outputDir, 'go', format, `${filename}-${timestamp}-${i}.${format}`);
            fs.writeFileSync(goFile, goRes.output);
            goSizeBytes = fs.statSync(goFile).size;
        }

        // Store result for report
        results.push({
            url,
            format,
            javascript: {
                status: jsRes.success ? 'success' : 'failed',
                executionTime: jsRes.executionTime,
                fileSize: jsSizeBytes,
                error: jsRes.error || null
            },
            go: {
                status: goRes.success ? 'success' : 'failed',
                executionTime: goRes.executionTime,
                fileSize: goSizeBytes,
                error: goRes.error || null
            }
        });
    }

    // Calculate averages
    const jsAvg = jsSuccess > 0 ? Math.round(jsTotal / jsSuccess) : 0;
    const goAvg = goSuccess > 0 ? Math.round(goTotal / goSuccess) : 0;

    console.log(`\n📊 ${format.toUpperCase()} Format Summary:`);
    console.log(`  JavaScript: ${jsSuccess}/${urls.length} success, ${jsAvg}ms average`);
    console.log(`  Go: ${goSuccess}/${urls.length} success, ${goAvg}ms average`);

    return {
        format,
        totalUrls: urls.length,
        javascript: {
            successful: jsSuccess,
            failed: urls.length - jsSuccess,
            totalTime: jsTotal,
            averageTime: jsAvg
        },
        go: {
            successful: goSuccess,
            failed: urls.length - goSuccess,
            totalTime: goTotal,
            averageTime: goAvg
        },
        results
    };
}

// Read URLs from file
function readUrls() {
    try {
        if (!fs.existsSync(CONFIG.testUrlsFile)) {
            throw new Error(`Test URLs file not found: ${CONFIG.testUrlsFile}`);
        }

        const content = fs.readFileSync(CONFIG.testUrlsFile, 'utf8');
        const urls = content
            .split('\n')
            .map(line => line.trim())
            .filter(line => line && line.startsWith('https://'));

        if (urls.length === 0) {
            throw new Error('No valid URLs found in test file');
        }

        return urls;
    } catch (error) {
        console.error('❌ Error reading URLs:', error.message);
        process.exit(1);
    }
}

// Main function
async function main() {
    console.log('🚀 Starting Parser Comparison Tests');
    console.log('====================================');

    const overallStart = performance.now();

    // Setup
    setupOutputDirectories();

    const parserInstalled = installPostlightParser();
    if (!parserInstalled) {
        console.error('❌ Cannot proceed without @postlight/parser');
        process.exit(1);
    }

    const goBuilt = buildGoParser();
    if (!goBuilt) {
        console.error('❌ Cannot proceed without Go parser');
        process.exit(1);
    }

    // Read URLs
    const urls = readUrls();
    console.log(`\nFound ${urls.length} URLs to test\n`);

    // Test both formats
    const jsonResults = await testFormat('json', urls);
    const markdownResults = await testFormat('markdown', urls);

    const overallEnd = performance.now();
    const totalTime = Math.round(overallEnd - overallStart);

    // Generate final report
    const report = {
        timestamp: new Date().toISOString(),
        totalExecutionTime: totalTime,
        testUrlsFile: CONFIG.testUrlsFile,
        totalUrls: urls.length,
        formats: {
            json: jsonResults,
            markdown: markdownResults
        }
    };

    // Save report
    fs.writeFileSync(CONFIG.comparisonReport, JSON.stringify(report, null, 2));

    // Summary
    console.log('\n🎯 Comparison Complete!');
    console.log('='.repeat(25));
    console.log(`Total time: ${(totalTime / 1000).toFixed(1)}s`);
    console.log(`JSON - JS: ${jsonResults.javascript.successful}/${urls.length}, Go: ${jsonResults.go.successful}/${urls.length}`);
    console.log(`Markdown - JS: ${markdownResults.javascript.successful}/${urls.length}, Go: ${markdownResults.go.successful}/${urls.length}`);
    console.log(`\n📁 Results saved to: ${CONFIG.outputDir}/`);
    console.log(`📊 Comparison report: ${CONFIG.comparisonReport}`);
}

// Handle errors gracefully
process.on('unhandledRejection', (error) => {
    console.error('❌ Unhandled error:', error.message);
    process.exit(1);
});

// Run main function
if (require.main === module) {
    main().catch(console.error);
}

module.exports = { main };