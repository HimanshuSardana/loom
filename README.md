# Loom — executable literate programming for Neovim

`.loom` documents combine prose, named code blocks, execution, tangling, and export.

## Install

```sh
go install github.com/HimanshuSardana/loom/cmd/loom@latest
```

## Quickstart

Create `fibonacci.loom`:

````markdown
# Fibonacci

```python {#fib} {run} {session=python}
def fib(n):
    if n <= 1:
        return n
    return fib(n-1) + fib(n-2)
```

```python {#calculate} {run} {session=python} {depends=fib}
[fib(i) for i in range(10)]
```
````

Then:

```sh
loom check fibonacci.loom
loom run fibonacci.loom --allow-execution
loom run fibonacci.loom --block calculate --allow-execution
loom tangle fibonacci.loom            # uses {file=} targets + [tangle] output
loom export html fibonacci.loom
loom export pdf fibonacci.loom        # needs `typst`
loom build fibonacci.loom --allow-execution
loom watch fibonacci.loom --allow-execution
loom clean fibonacci.loom
```

Security: code never executes on open. `loom run/build` prompts for
confirmation on a TTY; pass `--allow-execution` for CI/non-interactive use.

## Block syntax

    ```<language> {#<name>} <options>

- name: `{#[A-Za-z][A-Za-z0-9_.-]*}` unique per document
- references: `<<name>>` (expanded during tangle, nested, cycle-checked)
- `{run}` makes a block executable
- `{session=name}` shares state (default: `default:<lang>`); `{isolated}` runs fresh
- `{file="path"}` tangle target; `{depends=a,b}` explicit deps

## Neovim

```lua
-- lazy.nvim
{ "HimanshuSardana/loom", config = function()
  require("loom").setup({ executable = "loom" })
end }
```

Commands: `:LoomBuild :LoomCheck :LoomTangle :LoomRun :LoomRunBlock
:LoomRunSession :LoomRunDocument :LoomExportHtml :LoomExportPdf
:LoomPreview :LoomClean :LoomClearOutput :LoomGotoDefinition :LoomReferences`

Suggested keymaps:

```lua
vim.keymap.set("n", "<leader>lr", "<cmd>LoomRunBlock<cr>")
vim.keymap.set("n", "<leader>lR", "<cmd>LoomRunDocument<cr>")
vim.keymap.set("n", "<leader>lt", "<cmd>LoomTangle<cr>")
vim.keymap.set("n", "<leader>lb", "<cmd>LoomBuild<cr>")
```

## Layout

- `cmd/loom` — CLI
- `internal/parser, ast, tangle, runtime, execution, cache, export/{html,typst}, config, diagnostics`
- `loom.nvim` — Neovim frontend (talks to CLI only)
- `examples/` — sample `.loom` docs

## Cache

Results cached in `.loom/cache/` keyed by source+language+session+deps.
`loom clean` clears it.
