#!/usr/bin/env node

const fs = require("fs");
const path = require("path");
const { execSync, execFileSync } = require("child_process");
const { performance } = require("perf_hooks");

// Configuration
const CONFIG = {
	testUrlsFile: process.argv[2] || "./testurls.txt",
	outputDir: "./test-output",
	goBinary: "../bin/hermes",
	comparisonReport: "./test-output/comparison-report.json",
};

// Ensure output directories exist and are clean
function setupOutputDirectories() {
	console.log("🧹 Cleaning up and setting up output directories...");

	const dirs = [
		`${CONFIG.outputDir}/js/json`,
		`${CONFIG.outputDir}/js/markdown`,
		`${CONFIG.outputDir}/go/json`,
		`${CONFIG.outputDir}/go/markdown`,
	];

	// Remove all files from output directories
	try {
		if (fs.existsSync(CONFIG.outputDir)) {
			fs.rmSync(CONFIG.outputDir, { recursive: true, force: true });
		}

		// Recreate directory structure
		dirs.forEach((dir) => {
			fs.mkdirSync(dir, { recursive: true });
		});

		console.log("✅ Output directories cleaned and created");
	} catch (error) {
		console.error("❌ Error setting up directories:", error.message);
		process.exit(1);
	}
}

// Install dependencies declared by benchmark/package.json.
function installDependencies() {
	console.log("📦 Installing benchmark dependencies...");

	try {
		execSync("npm install", {
			stdio: "inherit",
			cwd: process.cwd(),
		});

		console.log("✅ Benchmark dependencies installed");
		return true;
	} catch (error) {
		console.error(
			"❌ Failed to install benchmark dependencies:",
			error.message,
		);
		return false;
	}
}

// Build Go parser
function buildGoParser() {
	console.log("🔨 Building Go parser...");

	try {
		const hermesDir = ".."; // Now we're in hermes/benchmark, so parent dir is hermes
		if (!fs.existsSync(hermesDir)) {
			throw new Error("hermes directory not found");
		}

		// Check if Makefile exists, otherwise use go build
		if (fs.existsSync(path.join(hermesDir, "Makefile"))) {
			execSync("make build", {
				cwd: hermesDir,
				stdio: "inherit",
			});
		} else {
			execSync("go build -o bin/hermes ./cmd/hermes", {
				cwd: hermesDir,
				stdio: "inherit",
			});
		}

		console.log("✅ Go parser built");
		return true;
	} catch (error) {
		console.error("❌ Failed to build Go parser:", error.message);
		return false;
	}
}

async function parseWithJS(url, format) {
	const Parser = require("@postlight/parser");
	const result = await Parser.parse(url, {
		contentType: format === "json" ? "html" : "markdown",
	});
	return format === "json" ? JSON.stringify(result, null, 2) : result.content || "";
}

function parseWithGo(url, format) {
	return execFileSync(
		CONFIG.goBinary,
		["parse", "--format", format, "--", url],
		{ encoding: "utf8", timeout: 30000 },
	).trim();
}

async function runParser(parse, format, urls) {
	const results = [];
	let totalTime = 0;
	let successful = 0;
	for (let i = 0; i < urls.length; i++) {
		process.stdout.write(`  URL ${i + 1}/${urls.length}... `);
		const startTime = performance.now();
		let result;
		try {
			const output = await parse(urls[i], format);
			result = {
				success: true,
				executionTime: Math.round(performance.now() - startTime),
				output,
			};
		} catch (error) {
			result = {
				success: false,
				executionTime: Math.round(performance.now() - startTime),
				error: error.message,
				output: "",
			};
		}
		results.push(result);
		totalTime += result.executionTime;
		if (result.success) {
			successful++;
			console.log(`✓ ${result.executionTime}ms`);
		} else {
			console.log(`✗ ${result.executionTime}ms (${result.error})`);
		}
	}
	return {
		results,
		stats: {
			successful,
			failed: urls.length - successful,
			totalTime,
			// Include failed attempts in both the total and the average.
			averageTime: urls.length > 0 ? Math.round(totalTime / urls.length) : 0,
		},
	};
}

function saveResult(result, parser, format, filename) {
	let fileSize = 0;
	if (result.success) {
		const file = path.join(CONFIG.outputDir, parser, format, filename);
		fs.writeFileSync(file, result.output);
		fileSize = fs.statSync(file).size;
	}
	return {
		status: result.success ? "success" : "failed",
		executionTime: result.executionTime,
		fileSize,
		error: result.error || null,
	};
}

// Sanitize filename for cross-platform compatibility
function sanitizeFilename(url) {
	return url
		.replace(/https?:\/\//, "")
		.replace(/[^a-zA-Z0-9.-]/g, "_")
		.substring(0, 100); // Limit length
}

// Test both parsers for a specific format
async function testFormat(format, urls) {
	console.log(
		`\n🚀 Testing ${format.toUpperCase()} format with ${urls.length} URLs`,
	);
	console.log("=".repeat(50));

	const results = [];
	const timestamp = new Date().toISOString().replace(/[:.]/g, "-");

	console.log("\n📦 JavaScript Parser - Processing all URLs...");
	const javascript = await runParser(parseWithJS, format, urls);
	console.log("\n🔧 Go Parser - Processing all URLs...");
	const go = await runParser(parseWithGo, format, urls);

	console.log("\n💾 Saving results and compiling report...");
	for (let i = 0; i < urls.length; i++) {
		const url = urls[i];
		const filename = `${sanitizeFilename(url)}-${timestamp}-${i}.${format}`;
		results.push({
			url,
			format,
			javascript: saveResult(javascript.results[i], "js", format, filename),
			go: saveResult(go.results[i], "go", format, filename),
		});
	}

	console.log(`\n📊 ${format.toUpperCase()} Format Summary:`);
	console.log(
		`  JavaScript: ${javascript.stats.successful}/${urls.length} success, ${javascript.stats.averageTime}ms average per attempt`,
	);
	console.log(`  Go: ${go.stats.successful}/${urls.length} success, ${go.stats.averageTime}ms average per attempt`);

	return {
		format,
		totalUrls: urls.length,
		javascript: javascript.stats,
		go: go.stats,
		results,
	};
}

// Read URLs from file
function readUrls() {
	try {
		if (!fs.existsSync(CONFIG.testUrlsFile)) {
			throw new Error(`Test URLs file not found: ${CONFIG.testUrlsFile}`);
		}

		const content = fs.readFileSync(CONFIG.testUrlsFile, "utf8");
		const urls = content
			.split("\n")
			.map((line) => line.trim())
			.filter((line) => line && line.startsWith("https://"));

		if (urls.length === 0) {
			throw new Error("No valid URLs found in test file");
		}

		return urls;
	} catch (error) {
		console.error("❌ Error reading URLs:", error.message);
		process.exit(1);
	}
}

// Main function
async function main() {
	console.log("🚀 Starting Parser Comparison Tests");
	console.log("====================================");

	const overallStart = performance.now();

	// Setup
	setupOutputDirectories();

	const dependenciesInstalled = installDependencies();
	if (!dependenciesInstalled) {
		console.error("❌ Cannot proceed without benchmark dependencies");
		process.exit(1);
	}

	const goBuilt = buildGoParser();
	if (!goBuilt) {
		console.error("❌ Cannot proceed without Go parser");
		process.exit(1);
	}

	// Read URLs
	const urls = readUrls();
	console.log(`\nFound ${urls.length} URLs to test\n`);

	// Test both formats
	const jsonResults = await testFormat("json", urls);
	const markdownResults = await testFormat("markdown", urls);

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
			markdown: markdownResults,
		},
	};

	// Save report
	fs.writeFileSync(CONFIG.comparisonReport, JSON.stringify(report, null, 2));

	// Summary
	console.log("\n🎯 Comparison Complete!");
	console.log("=".repeat(25));
	console.log(`Total time: ${(totalTime / 1000).toFixed(1)}s`);
	console.log(
		`JSON - JS: ${jsonResults.javascript.successful}/${urls.length}, Go: ${jsonResults.go.successful}/${urls.length}`,
	);
	console.log(
		`Markdown - JS: ${markdownResults.javascript.successful}/${urls.length}, Go: ${markdownResults.go.successful}/${urls.length}`,
	);
	console.log(`\n📁 Results saved to: ${CONFIG.outputDir}/`);
	console.log(`📊 Comparison report: ${CONFIG.comparisonReport}`);
}

// Handle errors gracefully
process.on("unhandledRejection", (error) => {
	console.error("❌ Unhandled error:", error.message);
	process.exit(1);
});

// Run main function
if (require.main === module) {
	main().catch(console.error);
}

module.exports = { main };
