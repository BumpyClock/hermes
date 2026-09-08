const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const test = require("node:test");

const source = fs.readFileSync(path.join(__dirname, "test-comparison.js"), "utf8");

function loadComparison({ go, js, times = [0, 10], files = new Map() }) {
	const sandbox = {
		require(name) {
			switch (name) {
				case "child_process":
					return {
						execSync() { throw new Error("Unexpected shell execution"); },
						execFileSync: go,
					};
				case "@postlight/parser":
					return { parse: js };
				case "perf_hooks":
					return { performance: { now: () => times.shift() } };
				case "fs":
					return {
						writeFileSync(file, output) { files.set(file, output); },
						statSync: (file) => ({ size: Buffer.byteLength(files.get(file)) }),
					};
				default:
					return require(name);
			}
		},
		Date: class extends Date {
			constructor() { super("2026-09-07T12:34:56.789Z"); }
		},
		module: { exports: {} },
		process: { argv: [], stdout: { write() {} }, on() {} },
		console: { log() {}, error() {} },
	};
	vm.runInNewContext(source, sandbox, { filename: "test-comparison.js" });
	return sandbox;
}

test("Go parser receives URL metacharacters as one literal argument", () => {
	const url = 'https://example.com/$(printf harmless)`printf literal`?q="a b"';
	let invocation;
	const comparison = loadComparison({
		go(file, args, options) {
			invocation = { file, args: Array.from(args), options };
			return "{}\n";
		},
	});
	assert.equal(comparison.parseWithGo(url, "json"), "{}");
	assert.equal(invocation.file, "../bin/hermes");
	assert.deepEqual(invocation.args, ["parse", "--format", "json", "--", url]);
	assert.ok(!invocation.options.shell, "must execute without a shell");
});

test("average latency includes failed attempts in both numerator and denominator", async () => {
	for (const allFailed of [false, true]) {
		let jsCalls = 0;
		let goCalls = 0;
		const comparison = loadComparison({
			times: [0, 10, 10, 110, 110, 120, 120, 220],
			js: async () => {
				if (allFailed || jsCalls++ > 0) throw new Error("fetch failed");
				return {};
			},
			go: () => {
				if (allFailed || goCalls++ > 0) throw new Error("fetch failed");
				return "{}";
			},
		});
		const result = await comparison.testFormat("json", ["https://example.com/a", "https://example.com/b"]);
		for (const stats of [result.javascript, result.go]) {
			assert.equal(stats.totalTime, 110);
			assert.equal(stats.averageTime, 55);
			assert.equal(stats.successful, allFailed ? 0 : 1);
			assert.equal(stats.failed, allFailed ? 2 : 1);
		}
	}
});

for (const format of ["json", "markdown"]) {
	test(`${format} report preserves schema, parser order, files, and mixed failures`, async () => {
		const urls = ["https://example.com/a", "https://example.com/b"];
		const calls = [];
		const files = new Map();
		const comparison = loadComparison({
			files,
			times: [0, 10, 10, 110, 110, 130, 130, 160],
			js: async (url, options) => {
				calls.push(["js", url, options.contentType]);
				if (url === urls[1]) throw new Error("JS failure");
				return { content: "é" };
			},
			go: (file, args) => {
				calls.push(["go", args[4], args[2]]);
				if (args[4] === urls[0]) throw new Error("Go failure");
				return format === "json" ? '  {"content":"ok"}\n' : "  Go text\n";
			},
		});
		const result = await comparison.testFormat(format, urls);
		const jsOutput = format === "json" ? '{\n  "content": "é"\n}' : "é";
		const goOutput = format === "json" ? '{"content":"ok"}' : "Go text";
		const expected = {
			format,
			totalUrls: 2,
			javascript: { successful: 1, failed: 1, totalTime: 110, averageTime: 55 },
			go: { successful: 1, failed: 1, totalTime: 50, averageTime: 25 },
			results: [
				{
					url: urls[0], format,
					javascript: { status: "success", executionTime: 10, fileSize: format === "json" ? 21 : 2, error: null },
					go: { status: "failed", executionTime: 20, fileSize: 0, error: "Go failure" },
				},
				{
					url: urls[1], format,
					javascript: { status: "failed", executionTime: 100, fileSize: 0, error: "JS failure" },
					go: { status: "success", executionTime: 30, fileSize: format === "json" ? 16 : 7, error: null },
				},
			],
		};
		assert.equal(JSON.stringify(result), JSON.stringify(expected));
		const contentType = format === "json" ? "html" : "markdown";
		assert.deepEqual(calls, [
			["js", urls[0], contentType], ["js", urls[1], contentType],
			["go", urls[0], format], ["go", urls[1], format],
		]);
		assert.deepEqual([...files], [
			[`test-output/js/${format}/example.com_a-2026-09-07T12-34-56-789Z-0.${format}`, jsOutput],
			[`test-output/go/${format}/example.com_b-2026-09-07T12-34-56-789Z-1.${format}`, goOutput],
		]);
	});
}

test("empty URL list produces zero totals and no files", async () => {
	const files = new Map();
	const comparison = loadComparison({ files });
	const result = await comparison.testFormat("json", []);
	assert.equal(JSON.stringify(result), JSON.stringify({
		format: "json",
		totalUrls: 0,
		javascript: { successful: 0, failed: 0, totalTime: 0, averageTime: 0 },
		go: { successful: 0, failed: 0, totalTime: 0, averageTime: 0 },
		results: [],
	}));
	assert.equal(files.size, 0);
});
