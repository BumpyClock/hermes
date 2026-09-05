const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const test = require("node:test");

const source = fs.readFileSync(path.join(__dirname, "test-comparison.js"), "utf8");

function loadComparison({ go, js, times = [0, 10] }) {
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
					return { writeFileSync() {}, statSync: () => ({ size: 2 }) };
				default:
					return require(name);
			}
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
	assert.equal(comparison.parseWithGo(url, "json").success, true);
	assert.equal(invocation.file, "../bin/hermes");
	assert.deepEqual(invocation.args, ["parse", "--format", "json", "--", url]);
	assert.notEqual(invocation.options.shell, true);
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
