const EXAMPLES = {
  hello: `// Hello, Lotus!
let greeting = "Hello from Lotus!"
print(greeting)
 
mut counter = 0
while counter < 5 {
    print("Count: " + str(counter))
    counter = counter + 1
}`,

  fibonacci: `// Fibonacci — recursive
fn fibonacci(n) {
    if n <= 1 { return n }
    return fibonacci(n - 1) + fibonacci(n - 2)
}
 
for i in range(0, 12) {
    print("fib(" + str(i) + ") = " + str(fibonacci(i)))
}`,

  closures: `// Closures & higher-order functions
fn make_counter(start) {
    mut n = start
    return fn() {
        n = n + 1
        return n
    }
}
 
let counter = make_counter(0)
print(str(counter()))   // 1
print(str(counter()))   // 2
print(str(counter()))   // 3
 
fn map(arr, f) {
    mut result = []
    for item in arr {
        result = push(result, f(item))
    }
    return result
}
 
let doubled = map([1, 2, 3, 4, 5], fn(x) { x * 2 })
print(str(doubled))`,

  classes: `// Classes & inheritance
class Animal {
    fn init(self, name, sound) {
        self.name = name
        self.sound = sound
    }
 
    fn speak(self) {
        return self.name + " says: " + self.sound
    }
 
    fn describe(self) {
        return "I am " + self.name
    }
}
 
class Dog extends Animal {
    fn init(self, name) {
        self.name = name
        self.sound = "Woof!"
    }
 
    fn fetch(self) {
        return self.name + " fetches the ball!"
    }
}
 
let cat = Animal("Whiskers", "Meow!")
let dog = Dog("Rex")
 
print(cat.speak())
print(dog.speak())
print(dog.fetch())`,

  sort: `// Quicksort implementation
fn quicksort(arr) {
    if len(arr) <= 1 { return arr }
 
    let pivot = arr[0]
    mut less    = []
    mut greater = []
 
    for i in range(1, len(arr)) {
        if arr[i] <= pivot {
            less = push(less, arr[i])
        } else {
            greater = push(greater, arr[i])
        }
    }
 
    mut result = quicksort(less)
    result = push(result, pivot)
    for item in quicksort(greater) {
        result = push(result, item)
    }
    return result
}
 
let data = [38, 27, 43, 3, 9, 82, 10, 55, 1, 99]
print("Input:  " + str(data))
print("Sorted: " + str(quicksort(data)))`,

  map: `// Maps and data structures
let person = {
    "name": "Alice",
    "age":  30,
    "city": "Luxembourg"
}
 
print("Name: " + person["name"])
print("Age:  " + str(person["age"]))
 
// Build a frequency map
fn freq(arr) {
    mut counts = {}
    for item in arr {
        let k = str(item)
        if counts[k] == nil {
            counts[k] = 1
        } else {
            counts[k] = counts[k] + 1
        }
    }
    return counts
}
 
let words = ["lotus", "go", "lotus", "vm", "go", "lotus"]
let counts = freq(words)
print("lotus appears " + str(counts["lotus"]) + " times")
print("go appears "    + str(counts["go"])    + " times")`,
};

const KEYWORDS_CTRL = new Set([
  "if",
  "else",
  "while",
  "for",
  "in",
  "return",
  "break",
  "continue",
  "import",
  "export",
  "from",
]);

const KEYWORDS_DECL = new Set(["let", "mut", "fn", "class", "extends"]);
const KEYWORDS_VAL = new Set(["true", "false", "nil"]);
const SELF_SUPER = new Set(["self", "super"]);
const BUILTINS = new Set([
  "print",
  "len",
  "push",
  "pop",
  "head",
  "tail",
  "type",
  "str",
  "int",
  "range",
]);
const PACKAGES = new Set(["Console", "Math", "OS"]);

// SYNTAX HIGHLIGHTING STUFF

function highlight(code) {
  // Tokenize line by line to handle comments correctly
  return code
    .split("\n")
    .map((line) => highlightLine(line))
    .join("\n");
}

function highlightLine(line) {
  // Strip comment first
  const commentIdx = findCommentStart(line);
  const code = commentIdx === -1 ? line : line.slice(0, commentIdx);
  const comment = commentIdx === -1 ? "" : line.slice(commentIdx);

  let result = highlightCode(code);
  if (comment) {
    result += `<span class="tok-comment">${esc(comment)}</span>`;
  }
  return result;
}

function findCommentStart(line) {
  let inStr = false;
  for (let i = 0; i < line.length - 1; i++) {
    if (line[i] === '"') {
      inStr = !inStr;
    }

    if (!inStr && line[i] === "/" && line[i + 1] === "/") {
      return i;
    }
  }

  return -1;
}

function highlightCode(code) {
  const tokens = tokenise(code);
  return tokens
    .map(([type, val]) => {
      if (type === "raw") {
        return esc(val);
      }
      return `<span class="${type}">${esc(val)}</span>`;
    })
    .join("");
}

function tokenise(code) {
  const tokens = [];
  let i = 0;
  while (i < code.length) {
    // String literal
    if (code[i] === '"') {
      let j = i + 1;
      while (j < code.length && !(code[j] === '"' && code[j - 1] !== "\\")) j++;
      tokens.push(["tok-str", code.slice(i, j + 1)]);
      i = j + 1;
      continue;
    }
    // Number
    if (/[0-9]/.test(code[i])) {
      let j = i;
      while (j < code.length && /[0-9.]/.test(code[j])) j++;
      tokens.push(["tok-num", code.slice(i, j)]);
      i = j;
      continue;
    }
    // Identifier / keyword
    if (/[a-zA-Z_]/.test(code[i])) {
      let j = i;
      while (j < code.length && /[a-zA-Z0-9_]/.test(code[j])) j++;
      const word = code.slice(i, j);
      let cls = "raw";
      if (KEYWORDS_CTRL.has(word)) cls = "tok-kw";
      else if (KEYWORDS_DECL.has(word)) cls = "tok-kw-decl";
      else if (KEYWORDS_VAL.has(word)) cls = "tok-bool";
      else if (word === "nil") cls = "tok-nil";
      else if (SELF_SUPER.has(word)) cls = "tok-self";
      else if (BUILTINS.has(word)) cls = "tok-builtin";
      else if (PACKAGES.has(word)) cls = "tok-pkg";
      else if (/^[A-Z]/.test(word)) cls = "tok-class";
      // Check if followed by ( → function name
      else {
        const after = code.slice(j).trimStart();
        if (after.startsWith("(")) cls = "tok-fn";
      }
      tokens.push([cls, word]);
      i = j;
      continue;
    }
    // Operators
    const twoChar = code.slice(i, i + 2);
    if (["==", "!=", "<=", ">=", "&&", "||", "++", "--"].includes(twoChar)) {
      tokens.push(["tok-op", twoChar]);
      i += 2;
      continue;
    }
    tokens.push(["raw", code[i]]);
    i++;
  }
  return tokens;
}

function esc(s) {
  return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}

// EDITOR SETUP
const codeInput = document.getElementById("codeInput");
const highlightLayer = document.getElementById("highlightLayer");
const lineNumbers = document.getElementById("lineNumbers");
const editorInner = document.getElementById("editorInner");
const editorScroll = document.getElementById("editorScroll");

function updateHighlight() {
  const code = codeInput.value;
  highlightLayer.innerHTML = highlight(code) + "\n"; // trailing \n keeps height in sync
  updateLineNumbers(code);
  updateCursor();
  syncScroll();
}

function updateLineNumbers(code) {
  const lines = (code.match(/\n/g) || []).length + 1;
  const current = lineNumbers.children.length;
  if (lines === current) {
    return;
  }

  lineNumbers.innerHTML = Array.from(
    { length: lines },
    (_, i) => `<div>${i + 1}</div>`,
  ).join("");
}

function updateCursor() {
  const pos = codeInput.selectionStart;
  const before = codeInput.value.slice(0, pos);
  const line = (before.match(/\n/g) || []).length + 1;
  const col = before.length - before.lastIndexOf("\n");
  document.getElementById("lineCol").textContent = `Ln ${line}, Col ${col}`;
}

function syncScroll() {
  highlightLayer.style.transform = `translate(-${editorScroll.scrollLeft}px, -${editorScroll.scrollTop}px)`;
  lineNumbers.style.transform = `translateY(-${editorScroll.scrollTop}px)`;
}

// Handle Tab key
codeInput.addEventListener("keydown", (e) => {
  if (e.key === "Tab") {
    e.preventDefault();
    const start = codeInput.selectionStart;
    const end = codeInput.selectionEnd;
    codeInput.value =
      codeInput.value.slice(0, start) + "    " + codeInput.value.slice(end);
    codeInput.selectionStart = codeInput.selectionEnd = start + 4;
    updateHighlight();
  }
  if ((e.ctrlKey || e.metaKey) && e.key === "Enter") {
    e.preventDefault();
    runCode();
  }
});

codeInput.addEventListener("input", updateHighlight);
codeInput.addEventListener("keyup", updateCursor);
codeInput.addEventListener("click", updateCursor);
editorScroll.addEventListener("scroll", syncScroll);

// Sync textarea height to content
function resizeTextarea() {
  codeInput.style.height = "auto";
  const minH = editorScroll.clientHeight;
  codeInput.style.height = Math.max(minH, codeInput.scrollHeight) + "px";
  editorInner.style.height = codeInput.style.height;
}

const ro = new ResizeObserver(resizeTextarea);
ro.observe(editorScroll);
codeInput.addEventListener("input", resizeTextarea);

// Output
const outputBody = document.getElementById("outputBody");
const outputIdle = document.getElementById("outputIdle");
const statusIndicator = document.getElementById("statusIndicator");
const statusText = document.getElementById("statusText");
const execTimeEl = document.getElementById("execTime");

function setStatus(state, text) {
  statusIndicator.className = "status-dot " + state;
  statusText.textContent = text;
}

function clearOutput() {
  outputBody.innerHTML = "";
  outputBody.appendChild(outputIdle);
  outputIdle.style.display = "flex";
  execTimeEl.textContent = "";
  setStatus("", "ready");
  document.getElementById("statusDot").style.background = "var(--muted)";
}

function appendLine(text, type = "stdout") {
  outputIdle.style.display = "none";
  const lines = String(text).split("\n");
  lines.forEach((content, i) => {
    const row = document.createElement("div");
    row.className = `output-line ${type}`;
    row.style.animationDelay = i * 20 + "ms";
    const prefix = document.createElement("span");
    prefix.className = "line-prefix";
    prefix.textContent = type === "stderr" ? "!" : type === "info" ? "#" : "›";
    const span = document.createElement("span");
    span.className = "line-content";
    span.textContent = content;
    row.appendChild(prefix);
    row.appendChild(span);
    outputBody.appendChild(row);
  });
  outputBody.scrollTop = outputBody.scrollHeight;
}

// ── Run ───────────────────────────────────────────────────
const runBtn = document.getElementById("runBtn");

async function runCode() {
  const code = codeInput.value.trim();
  if (!code) {
    return;
  }

  runBtn.classList.add("loading");
  runBtn.querySelector("svg").innerHTML =
    '<circle cx="6" cy="6" r="4" stroke="currentColor" stroke-width="1.5" fill="none" stroke-dasharray="20" stroke-dashoffset="20" style="animation:dash 0.8s linear infinite"/>';
  clearOutput();
  setStatus("run", "running…");

  const t0 = performance.now();

  try {
    const resp = await fetch("/run", {
      method: "POST",
      headers: { "Content-Type": "text/plain" },
      body: code,
    });

    const elapsed = ((performance.now() - t0) / 1000).toFixed(3);
    execTimeEl.textContent = `${elapsed}s`;

    const data = await resp.json();

    if (data.stdout && data.stdout.trim()) {
      appendLine(data.stdout.trimEnd(), "stdout");
    }
    if (data.stderr && data.stderr.trim()) {
      appendLine(data.stderr.trimEnd(), "stderr");
    }
    if (!data.stdout && !data.stderr) {
      appendLine("(no output)", "info");
    }

    if (data.error) {
      appendLine(data.error, "stderr");
      setStatus("err", "error");
      document.getElementById("statusDot").style.background = "var(--red)";
    } else {
      setStatus("ok", `done in ${elapsed}s`);
      document.getElementById("statusDot").style.background = "var(--green)";
    }
  } catch (err) {
    appendLine(
      "Could not reach the Lotus server.\nMake sure lotus --playground is running.",
      "stderr",
    );
    setStatus("err", "connection error");
    document.getElementById("statusDot").style.background = "var(--red)";
  } finally {
    runBtn.classList.remove("loading");
    runBtn.querySelector("svg").innerHTML = '<polygon points="2,1 11,6 2,11"/>';
  }
}

runBtn.addEventListener("click", runCode);
document.getElementById("clearBtn").addEventListener("click", clearOutput);

//  Examples
document.querySelectorAll(".example-btn").forEach((btn) => {
  btn.addEventListener("click", () => {
    const code = EXAMPLES[btn.dataset.example];
    if (code) {
      codeInput.value = code;
      updateHighlight();
      resizeTextarea();
      clearOutput();
      codeInput.focus();
    }
  });
});

// Init
codeInput.value = EXAMPLES.hello;
updateHighlight();
resizeTextarea();

// Add CSS keyframe for loading spinner via JS
const style = document.createElement("style");
style.textContent = `@keyframes dash { to { stroke-dashoffset: 0; } }`;
document.head.appendChild(style);
