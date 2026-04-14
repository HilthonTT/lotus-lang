const EXAMPLES = {
  hello: `// Hello, Lotus!
let name = "Lotus"
let version = 1
print("Hello from \${name} v\${version}!")

mut counter = 0
while counter < 5 {
    print("Count: \${counter}")
    counter++
}`,

  fibonacci: `// Fibonacci — recursive
fn fibonacci(n: int) -> int {
    if n <= 1 { return n }
    return fibonacci(n - 1) + fibonacci(n - 2)
}

for i in range(0, 12) {
    print("fib(\${i}) = \${fibonacci(i)}")
}`,

  closures: `// Closures & higher-order functions
fn make_counter(start: int) {
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

let doubled = Array.map([1, 2, 3, 4, 5], fn(x) { x * 2 })
print(str(doubled))`,

  classes: `// Classes & inheritance
class Animal {
    fn init(self, name: string, sound: string) {
        self.name = name
        self.sound = sound
    }

    fn speak(self) -> string {
        return "\${self.name} says: \${self.sound}"
    }
}

class Dog extends Animal {
    fn init(self, name: string) {
        self.name = name
        self.sound = "Woof!"
    }

    fn fetch(self) -> string {
        return "\${self.name} fetches the ball!"
    }
}

let cat = Animal("Whiskers", "Meow!")
let dog = Dog("Rex")

print(cat.speak())
print(dog.speak())
print(dog.fetch())`,

  sort: `// Quicksort implementation
fn quicksort(arr: array) -> array {
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
print("Input:  \${str(data)}")
print("Sorted: \${str(quicksort(data))}")`,

  map: `// Maps and data structures
let person = {
    "name": "Alice",
    "age":  30,
    "city": "Luxembourg"
}

print("Name: \${person["name"]}")
print("Age:  \${str(person["age"])}")

fn freq(arr) {
    mut counts = {}
    for item in arr {
        let k = str(item)
        counts[k] = (counts[k] ?? 0) + 1
    }
    return counts
}

let words = ["lotus", "go", "lotus", "vm", "go", "lotus"]
let counts = freq(words)
print("lotus appears \${str(counts["lotus"])} times")
print("go appears \${str(counts["go"])} times")`,

  match: `// Match expressions & Enums
enum Direction { North, South, East, West }
enum Shape { Circle(radius), Rect(width, height) }

let dir = Direction.North
let label = match dir {
    Direction.North -> "heading north",
    Direction.South -> "heading south",
    _ -> "going somewhere"
}
print(label)

let s = Shape.Circle(5.0)
print(str(s))

fn grade(score: int) -> string {
    return match score {
        100 -> "A+",
        90  -> "A",
        80  -> "B",
        70  -> "C",
        _   -> "F"
    }
}
print(grade(90))
print(grade(75))`,

  strings: `// String package
let s = "  Hello, Lotus World!  "
print(String.trim(s))
print(String.upper("lotus"))
print(String.lower("LOTUS"))
print(String.replace(s, "World", "Universe"))
print(str(String.contains(s, "Lotus")))
print(str(String.split("a,b,c,d", ",")))
print(String.repeat("ab", 3))
print(String.padLeft("7", 3, "0"))

let name = "anonymous"
let lang = "Lotus"
print("Hello \${name}, welcome to \${lang}!")
print("2 + 2 = \${2 + 2}")`,

  arrayFns: `// Array package — higher-order functions
let nums = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]

let evens   = Array.filter(nums, fn(x) { x % 2 == 0 })
let doubled = Array.map(nums, fn(x) { x * 2 })
let sum     = Array.reduce(nums, fn(acc, x) { acc + x }, 0)
let first   = Array.find(nums, fn(x) { x > 5 })

print("Evens:   \${str(evens)}")
print("Doubled: \${str(doubled)}")
print("Sum:     \${str(sum)}")
print("First>5: \${str(first)}")

let sorted = Array.sort([3, 1, 4, 1, 5, 9, 2, 6])
print("Sorted:  \${str(sorted)}")
print("Unique:  \${str(Array.unique([1,1,2,2,3]))}")
print("Any>9:   \${str(Array.any(nums, fn(x) { x > 9 }))}")
print("All>0:   \${str(Array.all(nums, fn(x) { x > 0 }))}")`,
};

// ── Token sets ──────────────────────────────────────────────
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
  "match",
]);
const KEYWORDS_DECL = new Set(["let", "mut", "fn", "class", "extends", "enum"]);
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
const PACKAGES = new Set([
  "Console",
  "Math",
  "OS",
  "Http",
  "Task",
  "String",
  "Array",
]);
const TYPES = new Set(["int", "float", "string", "bool", "array", "map"]);

// ── Syntax highlighter ──────────────────────────────────────
function highlight(code) {
  return code.split("\n").map(highlightLine).join("\n");
}

function highlightLine(line) {
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
  return tokenise(code)
    .map(([type, val]) =>
      type === "raw" ? esc(val) : `<span class="${type}">${esc(val)}</span>`,
    )
    .join("");
}

function tokenise(code) {
  const tokens = [];
  let i = 0;
  while (i < code.length) {
    // String literal — handles ${...} interpolation highlighting
    if (code[i] === '"') {
      let j = i + 1;
      let buf = '"';
      while (j < code.length && code[j] !== '"') {
        if (code[j] === "\\") {
          buf += code[j] + (code[j + 1] || "");
          j += 2;
          continue;
        }
        // Highlight ${expr} inside string
        if (code[j] === "$" && code[j + 1] === "{") {
          if (buf.length > 1) {
            tokens.push(["tok-str", buf]);
          }
          buf = "";
          tokens.push(["tok-op", "${"]);
          j += 2;
          let depth = 1,
            exprBuf = "";
          while (j < code.length && depth > 0) {
            if (code[j] === "{") depth++;
            if (code[j] === "}") {
              depth--;
              if (depth === 0) {
                j++;
                break;
              }
            }
            exprBuf += code[j++];
          }
          // Recursively highlight the expression inside
          if (exprBuf) {
            tokenise(exprBuf).forEach((t) => tokens.push(t));
          }
          tokens.push(["tok-op", "}"]);
          continue;
        }
        buf += code[j++];
      }
      buf += j < code.length ? '"' : "";
      if (buf) tokens.push(["tok-str", buf]);
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

      if (KEYWORDS_CTRL.has(word)) {
        cls = "tok-kw";
      } else if (KEYWORDS_DECL.has(word)) {
        cls = "tok-kw-decl";
      } else if (KEYWORDS_VAL.has(word)) {
        cls = "tok-bool";
      } else if (SELF_SUPER.has(word)) {
        cls = "tok-self";
      } else if (BUILTINS.has(word)) {
        cls = "tok-builtin";
      } else if (PACKAGES.has(word)) {
        cls = "tok-pkg";
      } else if (TYPES.has(word)) {
        // Only colour as type if preceded by ': ' or '->'
        const before = code.slice(0, i).trimEnd();
        if (before.endsWith(":") || before.endsWith("->")) {
          cls = "tok-type";
        }
      } else if (/^[A-Z]/.test(word)) {
        cls = "tok-class";
      } else {
        const after = code.slice(j).trimStart();
        if (after.startsWith("(")) {
          cls = "tok-fn";
        }
      }
      tokens.push([cls, word]);
      i = j;
      continue;
    }

    // Multi-char operators (longest match first)
    const threeChar = code.slice(i, i + 3);
    const twoChar = code.slice(i, i + 2);

    if (["<<=", ">>="].includes(threeChar)) {
      tokens.push(["tok-op", threeChar]);
      i += 3;
      continue;
    }
    if (
      [
        "==",
        "!=",
        "<=",
        ">=",
        "&&",
        "||",
        "++",
        "--",
        "??",
        "?.",
        "->",
        "<<",
        ">>",
      ].includes(twoChar)
    ) {
      tokens.push(["tok-op", twoChar]);
      i += 2;
      continue;
    }
    if ("&|^~".includes(code[i])) {
      tokens.push(["tok-op", code[i]]);
      i++;
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

// ── Editor setup ────────────────────────────────────────────
const codeInput = document.getElementById("codeInput");
const highlightLayer = document.getElementById("highlightLayer");
const lineNumbers = document.getElementById("lineNumbers");
const editorInner = document.getElementById("editorInner");
const editorScroll = document.getElementById("editorScroll");

function updateHighlight() {
  const code = codeInput.value;
  highlightLayer.innerHTML = highlight(code) + "\n";
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

function resizeTextarea() {
  codeInput.style.height = "auto";
  const minH = editorScroll.clientHeight;
  codeInput.style.height = Math.max(minH, codeInput.scrollHeight) + "px";
  editorInner.style.height = codeInput.style.height;
}
const ro = new ResizeObserver(resizeTextarea);
ro.observe(editorScroll);
codeInput.addEventListener("input", resizeTextarea);

// ── Output ──────────────────────────────────────────────────
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

// ── Run ─────────────────────────────────────────────────────
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

// ── Examples ────────────────────────────────────────────────
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

// ── Init ────────────────────────────────────────────────────
codeInput.value = EXAMPLES.hello;
updateHighlight();
resizeTextarea();

const style = document.createElement("style");
style.textContent = `
  @keyframes dash { to { stroke-dashoffset: 0; } }
  .tok-type { color: #a8d8a8; font-style: italic; }
`;
document.head.appendChild(style);
