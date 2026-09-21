if vim.g.loaded_loom == 1 then finish endif
let g:loaded_loom = 1
" loom.nvim commands are created by lua setup(); filetype detection:
augroup loom_ft
  autocmd!
  autocmd BufNewFile,BufRead *.loom set filetype=loom
augroup END
