local M = {}

M.config = { executable = "loom", auto_diag = true }
local ns = vim.api.nvim_create_namespace("loom_output")
local diag_ns = vim.api.nvim_create_namespace("loom_diag")

function M.setup(opts)
  M.config = vim.tbl_extend("force", M.config, opts or {})
  local cmds = {
    LoomBuild = "build", LoomCheck = "check", LoomTangle = "tangle",
    LoomRun = "run", LoomRunBlock = "runblock", LoomRunSession = "runsession",
    LoomRunDocument = "rundoc", LoomExportHtml = "exporthtml",
    LoomExportPdf = "exportpdf", LoomPreview = "preview", LoomClean = "clean",
  }
  for name, _ in pairs(cmds) do
    vim.api.nvim_create_user_command(name, function(o) M.dispatch(name, o) end, { nargs = "*" })
  end
  vim.api.nvim_create_user_command("LoomClearOutput", function() M.clear() end, {})
  vim.api.nvim_create_user_command("LoomGotoDefinition", function() M.goto_def() end, {})
  vim.api.nvim_create_user_command("LoomReferences", function() M.references() end, {})
end

local function loom_file()
  local f = vim.api.nvim_buf_get_name(0)
  if f == "" then vim.notify("loom: save buffer first", vim.log.levels.ERROR) return nil end
  return f
end

local function run_cli(args, on_exit)
  vim.fn.jobstart(args, {
    stdout_buffered = true, stderr_buffered = true,
    on_stdout = function(_, d) if on_exit then on_exit.out = ((on_exit.out or "") .. table.concat(d, "\n")) end end,
    on_stderr = function(_, d) if on_exit then on_exit.err = ((on_exit.err or "") .. table.concat(d, "\n")) end end,
    on_exit = function(_, code)
      vim.schedule(function() on_exit.cb(code, on_exit.out or "", on_exit.err or "") end)
    end,
  })
end

-- Find fenced code block containing cursor. Returns {name, lang, start, finish}.
function M.current_block()
  local cur = vim.api.nvim_win_get_cursor(0)[1]
  local lines = vim.api.nvim_buf_get_lines(0, 0, -1, false)
  local open = nil
  for i, line in ipairs(lines) do
    if line:match("^%s*```") then
      if open == nil then
        local lang = line:match("^%s*```(%S*)") or ""
        local name = line:match("{#([A-Za-z][A-Za-z0-9_%.%-]*)}")
        open = { lang = lang, name = name, start = i }
      else
        if open.start <= cur and cur <= i then
          open.finish = i
          return open
        end
        open = nil
      end
    end
  end
  return nil
end

function M.show_output(block, text)
  M.clear()
  if not block or not text or text == "" then return end
  local line = block.finish - 1
  vim.api.nvim_buf_set_extmark(0, ns, line, 0, {
    virt_lines = vim.tbl_map(function(l) return { { "→ " .. l, "Comment" } } end, vim.split(text, "\n")),
    virt_lines_above = false,
  })
end

function M.clear()
  vim.api.nvim_buf_clear_namespace(0, ns, 0, -1)
end

function M.set_diagnostics(diag_text)
  local qf = {}
  for line in (diag_text or ""):gmatch("[^\n]+") do
    local f, l, c, m = line:match("^(.-):(%d+):(%d+)%s*(.*)$")
    if f then
      table.insert(qf, { filename = f, lnum = tonumber(l), col = tonumber(c), text = m })
    end
  end
  if #qf > 0 then
    vim.fn.setqflist(qf)
    vim.diagnostic.set(diag_ns, 0, vim.tbl_map(function(e)
      return { lnum = e.lnum - 1, col = e.col - 1, message = e.text, severity = vim.diagnostic.severity.ERROR }
    end, qf))
  end
end

function M.dispatch(cmd, _)
  local f = loom_file()
  if not f then return end
  local exe = M.config.executable
  if cmd == "LoomRunBlock" then
    local b = M.current_block()
    if not b or not b.name then
      vim.notify("loom: cursor is not inside a named code block", vim.log.levels.ERROR)
      return
    end
    run_cli({ exe, "run", f, "--block", b.name, "--allow-execution", "--json" }, {
      cb = function(code, out, err)
        if code ~= 0 then
          vim.notify("loom run failed:\n" .. err .. out, vim.log.levels.ERROR)
          M.set_diagnostics(err .. "\n" .. out)
          return
        end
        local ok, data = pcall(vim.json.decode, out)
        if ok and data and data[1] then
          local r = data[#data]
          local txt = (r.stdout or "") .. (r.value ~= "" and ("=> " .. r.value .. "\n") or "") .. (r.stderr or "")
          M.show_output(b, vim.trim(txt))
          vim.notify("loom: block " .. b.name .. " ok")
        else
          vim.notify(out, vim.log.levels.INFO)
        end
      end,
    })
  elseif cmd == "LoomRun" or cmd == "LoomRunDocument" then
    run_cli({ exe, "run", f, "--allow-execution" }, { cb = function(c, o, e)
      if c ~= 0 then vim.notify("loom:\n" .. e .. o, vim.log.levels.ERROR) M.set_diagnostics(e) else vim.notify(o, vim.log.levels.INFO) end
    end })
  elseif cmd == "LoomRunSession" then
    local b = M.current_block()
    local extra = {}
    if b and b.name then extra = { "--block", b.name } end
    run_cli(vim.list_extend({ exe, "run", f, "--allow-execution" }, extra), { cb = function(c, o, e)
      if c ~= 0 then vim.notify("loom:\n" .. e .. o, vim.log.levels.ERROR) else vim.notify(o, vim.log.levels.INFO) end
    end })
  elseif cmd == "LoomCheck" then
    run_cli({ exe, "check", f }, { cb = function(c, o, e)
      if c ~= 0 then vim.notify("loom check:\n" .. e, vim.log.levels.ERROR) M.set_diagnostics(e) else vim.notify("loom: ok", vim.log.levels.INFO) end
    end })
  elseif cmd == "LoomBuild" then
    run_cli({ exe, "build", f, "--allow-execution" }, { cb = function(c, o, e)
      if c ~= 0 then vim.notify("loom build failed:\n" .. e .. o, vim.log.levels.ERROR) M.set_diagnostics(e .. o) else vim.notify("loom build complete", vim.log.levels.INFO) end
    end })
  elseif cmd == "LoomTangle" then
    run_cli({ exe, "tangle", f }, { cb = function(c, o, e)
      if c ~= 0 then vim.notify(e, vim.log.levels.ERROR) else vim.notify(o, vim.log.levels.INFO) end
    end })
  elseif cmd == "LoomExportHtml" then
    run_cli({ exe, "export", "html", f }, { cb = function(c, o, e)
      if c ~= 0 then vim.notify(e, vim.log.levels.ERROR) else vim.notify(o, vim.log.levels.INFO) end
    end })
  elseif cmd == "LoomExportPdf" then
    run_cli({ exe, "export", "pdf", f }, { cb = function(c, o, e)
      if c ~= 0 then vim.notify(e, vim.log.levels.ERROR) else vim.notify(o, vim.log.levels.INFO) end
    end })
  elseif cmd == "LoomPreview" then
    run_cli({ exe, "export", "html", f }, { cb = function(c, o, e)
      if c ~= 0 then vim.notify(e, vim.log.levels.ERROR) end
    end })
  elseif cmd == "LoomClean" then
    run_cli({ exe, "clean", f }, { cb = function(c, o, _) vim.notify("loom clean done") end })
  end
end

function M.goto_def()
  local word = vim.fn.expand("<cWORD>")
  local ref = word:match("<<([A-Za-z][A-Za-z0-9_%.%-]*)>>") or vim.fn.expand("<cword>")
  if not ref or ref == "" then vim.notify("loom: no reference under cursor", vim.log.levels.WARN) return end
  local lines = vim.api.nvim_buf_get_lines(0, 0, -1, false)
  for i, line in ipairs(lines) do
    if line:match("{#" .. vim.pesc(ref) .. "}") then
      vim.api.nvim_win_set_cursor(0, { i, 0 })
      vim.cmd("normal! zz")
      return
    end
  end
  vim.notify("loom: definition not found: " .. ref, vim.log.levels.ERROR)
end

function M.references()
  local b = M.current_block()
  local name = b and b.name
  if not name then vim.notify("loom: cursor not in a named block", vim.log.levels.WARN) return end
  local lines = vim.api.nvim_buf_get_lines(0, 0, -1, false)
  local qf = {}
  for i, line in ipairs(lines) do
    if line:find("<<" .. name .. ">>", 1, true) then
      table.insert(qf, { bufnr = vim.api.nvim_get_current_buf(), lnum = i, col = 1, text = line })
    end
  end
  vim.fn.setqflist(qf)
  vim.cmd("copen")
end

return M
