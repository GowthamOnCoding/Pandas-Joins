" Add runtime paths for plugins
set runtimepath+=C:\Users\meena\.vim\pack\plugins\start\nerdtree
set runtimepath+=C:\Users\meena\.vim\pack\plugins\start\vim-fugitive
set runtimepath+=C:\Users\meena\.vim\pack\plugins\start\vim-airline
set runtimepath+=C:\Users\meena\.vim\pack\plugins\start\vim-airline-themes
set runtimepath+=C:\Users\meena\.vim\pack\plugins\start\vim-gitgutter
set runtimepath+=C:\Users\meena\.vim\pack\plugins\start\grep.vim
set runtimepath+=C:\Users\meena\.vim\pack\plugins\start\CSApprox
set runtimepath+=C:\Users\meena\.vim\pack\plugins\start\delimitMate
set runtimepath+=C:\Users\meena\.vim\pack\plugins\start\tagbar
set runtimepath+=C:\Users\meena\.vim\pack\plugins\start\ale
set runtimepath+=C:\Users\meena\.vim\pack\plugins\start\material.vim
set runtimepath+=C:\Users\meena\.vim\pack\plugins\start\fzf
set runtimepath+=C:\Users\meena\.vim\pack\plugins\start\fzf.vim
set runtimepath+=C:\Users\meena\.vim\pack\plugins\start\vim-misc
set runtimepath+=C:\Users\meena\.vim\pack\plugins\start\vim-session
set runtimepath+=C:\Users\meena\.vim\pack\plugins\start\vim-snippets
set runtimepath+=C:\Users\meena\.vim\pack\plugins\start\vim-go


" Basic Setup
set nocompatible
filetype off
set encoding=utf-8
set fileencoding=utf-8
set fileencodings=utf-8
set ttyfast
set backspace=indent,eol,start
set tabstop=4
set softtabstop=0
set shiftwidth=4
set expandtab
let mapleader=','
set hidden
set hlsearch
set incsearch
set ignorecase
set smartcase
set fileformats=unix,dos,mac
if exists('$SHELL')
    set shell=$SHELL
else
    set shell=/bin/sh
endif
let g:session_directory = "~/.vim/session"
let g:session_autoload = "no"
let g:session_autosave = "no"
let g:session_command_aliases = 1
syntax on
set ruler
set number
let no_buffers_menu=1
set wildmenu
set mouse=a
set mousemodel=popup
set t_Co=256
set guioptions=egmrti
set gfn=Monospace\ 10
if has("gui_running")
  if has("gui_mac") || has("gui_macvim")
    set guifont=Menlo:h12
    set transparency=7
  endif
else
  let g:CSApprox_loaded = 1
  let g:indentLine_enabled = 1
  let g:indentLine_concealcursor = ''
  let g:indentLine_char = '┆'
  let g:indentLine_faster = 1
  if $COLORTERM == 'gnome-terminal'
    set term=gnome-256color
  else
    if $TERM == 'xterm'
      set term=xterm-256color
    endif
  endif
endif
if &term =~ '256color'
  set t_ut=
endif
set gcr=a:blinkon0
set scrolloff=3
set laststatus=2
set modeline
set modelines=10
set title
set titleold="Terminal"
set titlestring=%F
set statusline=%F%m%r%h%w%=(%{&ff}/%Y)\ (line\ %l\/%L,\ col\ %c)\

" Search mappings
nnoremap n nzzzv
nnoremap N Nzzzv

" Abbreviations
cnoreabbrev W! w!
cnoreabbrev Q! q!
cnoreabbrev Qall! qall!
cnoreabbrev Wq wq
cnoreabbrev Wa wa
cnoreabbrev wQ wq
cnoreabbrev WQ wq
cnoreabbrev W w
cnoreabbrev Q q
cnoreabbrev Qall qall

" Terminal emulation
nnoremap <silent> <leader>sh :terminal<CR>

" Commands
command! FixWhitespace :%s/\s\+$//e

" Functions
if !exists('*s:setupWrapping')
  function s:setupWrapping()
    set wrap
    set wm=2
    set textwidth=79
  endfunction
endif

" Autocmd Rules
augroup vimrc-sync-fromstart
  autocmd!
  autocmd BufEnter * :syntax sync maxlines=200
augroup END
augroup vimrc-remember-cursor-position
  autocmd!
  autocmd BufReadPost * if line("'\"") > 1 && line("'\"") <= line("$") | exe "normal! g`\"" | endif
augroup END
augroup vimrc-wrapping
  autocmd!
  autocmd BufRead,BufNewFile *.txt call s:setupWrapping()
augroup END
augroup vimrc-make-cmake
  autocmd!
  autocmd FileType make setlocal noexpandtab
  autocmd BufNewFile,BufRead CMakeLists.txt setlocal filetype=cmake
augroup END

set autoread

" Mappings
noremap <Leader>h :<C-u>split<CR>
noremap <Leader>v :<C-u>vsplit<CR>
noremap <Leader>ga :Gwrite<CR>
noremap <Leader>gc :Git commit --verbose<CR>
noremap <Leader>gsh :Git push<CR>
noremap <Leader>gll :Git pull<CR>
noremap <Leader>gs :Git<CR>
noremap <Leader>gb :Git blame<CR>
noremap <Leader>gd :Gvdiffsplit<CR>
noremap <Leader>gr :GRemove<CR>
nnoremap <leader>so :OpenSession<Space>
nnoremap <leader>ss :SaveSession<Space>
nnoremap <leader>sd :DeleteSession<CR>
nnoremap <leader>sc :CloseSession<CR>
nnoremap <Tab> gt
nnoremap <S-Tab> gT
nnoremap <silent> <S-t> :tabnew<CR>
nnoremap <leader>. :lcd %:p:h<CR>
noremap <Leader>e :e <C-R>=expand("%:p:h") . "/" <CR>
noremap <Leader>te :tabe <C-R>=expand("%:p:h") . "/" <CR>
set wildmode=list:longest,list:full
set wildignore+=*.o,*.obj,.git,*.rbc,*.pyc,__pycache__
let $FZF_DEFAULT_COMMAND =  "find * -path '*/\.*' -prune -o -path 'node_modules/**' -prune -o -path 'target/**' -prune -o -path 'dist/**' -prune -o  -type f -print -o -type l -print 2> /dev/null"
if executable('ag')
  let $FZF_DEFAULT_COMMAND = 'ag --hidden --ignore .git -g ""'
  set grepprg=ag\ --nogroup\ --nocolor
endif
if executable('rg')
  let $FZF_DEFAULT_COMMAND = 'rg --files --hidden --follow --glob "!.git/*"'
  set grepprg=rg\ --vimgrep
  command! -bang -nargs=* Find call fzf#vim#grep('rg --column --line-number --no-heading --fixed-strings --ignore-case --hidden --follow --glob "!.git/*" --color "always" '.shellescape(<q-args>).'| tr -d "\017"', 1, <bang>0)
endif
cnoremap <C-P> <C-R>=expand("%:p:h") . "/" <CR>
nnoremap <silent> <leader>b :Buffers<CR>
nnoremap <silent> <leader>e :FZF -m<CR>
nmap <leader>y :History:<CR>

" Basic settings for vim-go
let g:go_def_mode = 'gopls'
let g:go_auto_sameids = 1
let g:go_fmt_command = "goimports"
let g:go_fmt_autosave = 1

" Enable auto-completion
let g:go_autocomplete = 1

" Enable go doc popup
let g:go_doc_popup_window = 1
let g:go_doc_keywordprg_enabled = 1

" Enable real-time documentation display
"autocmd CursorHold * silent call go#doc#Definition()

" Key mapping to show Go documentation
"nnoremap <silent> K :GoDoc<CR>

" vim-go key mappings
autocmd FileType go nmap <leader>b :GoBuild<CR>
autocmd FileType go nmap <leader>r :GoRun<CR>
autocmd FileType go nmap <leader>t :GoTest<CR>
autocmd FileType go nmap <leader>c :GoCoverageToggle<CR>
autocmd FileType go nmap <leader>i :GoImpl<CR>
autocmd FileType go nmap <leader>g :GoGenerate<CR>
autocmd FileType go nmap <leader>m :GoMetaLinter<CR>
autocmd FileType go nmap <leader>d :GoDoc<CR>
autocmd FileType go nmap <leader>f :GoFmt<CR>

" ALE configuration
let g:ale_linters = {}
:call extend(g:ale_linters, {
    \"go": ['golint', 'go vet'], })

" Tagbar configuration
nmap <silent> <F4> :TagbarToggle<CR>
let g:tagbar_autofocus = 1

" Disable visualbell
set noerrorbells visualbell t_vb=
if has('autocmd')
  autocmd GUIEnter * set visualbell t_vb=
endif

" Copy/Paste/Cut
if has('unnamedplus')
  set clipboard=unnamed,unnamedplus
endif
noremap YY "+y<CR>
noremap <leader>p "+gP[_{{{CITATION{{{_1{](https://github.com/wborbajr/dotfiles/tree/99f48b8b40e65fa46c803890f4137b276fc3dc03/nvim%2Finit.vim.GO)
" Copy/Paste/Cut (continued)
if has('unnamedplus')
  set clipboard=unnamed,unnamedplus
endif
noremap YY "+y<CR>
noremap <leader>p "+gP<CR>
noremap XX "+x<CR>
if has('macunix')
  " pbcopy for OSX copy/paste
  vmap <C-x> :!pbcopy<CR>
  vmap <C-c> :w !pbcopy<CR><CR>
endif

" Buffer navigation
noremap <leader>z :bp<CR>
noremap <leader>x :bn<CR>
noremap <leader>w :bn<CR>

" Close buffer
noremap <leader>c :bd<CR>

" Clean search (highlight)
nnoremap <silent> <leader><space> :noh<cr>

" Switching windows
noremap <C-j> <C-w>j
noremap <C-k> <C-w>k
noremap <C-l> <C-w>l
noremap <C-h> <C-w>h

" Vmap for maintaining Visual Mode after shifting > and <
vmap < <gv
vmap > >gv

" Move visual block
vnoremap J :m '>+1<CR>gv=gv
vnoremap K :m '<-2<CR>gv=gv

" Open current line on GitHub
nnoremap <Leader>o :.GBrowse<CR>

" Custom configs for vim-go
function! s:build_go_files()
  let l:file = expand('%')
  if l:file =~# '^\f\+_test\.go$'
    call go#test#Test(0, 1)
  elseif l:file =~# '^\f\+\.go$'
    call go#cmd#Build(0)
  endif
endfunction

let g:go_list_type = "quickfix"
let g:go_fmt_command = "goimports"
let g:go_fmt_fail_silently = 1

let g:go_highlight_types = 1
let g:go_highlight_fields = 1
let g:go_highlight_functions = 1
let g:go_highlight_methods = 1
let g:go_highlight_operators = 1
let g:go_highlight_build_constraints = 1
let g:go_highlight_structs = 1
let g:go_highlight_generate_tags = 1
let g:go_highlight_space_tab_error = 0
let g:go_highlight_array_whitespace_error = 0
let g:go_highlight_trailing_whitespace_error = 0
let g:go_highlight_extra_types = 1

autocmd BufNewFile,BufRead *.go setlocal noexpandtab tabstop=4 shiftwidth=4 softtabstop=4

augroup completion_preview_close
  autocmd!
  if v:version > 703 || v:version == 703 && has('patch598')
    autocmd CompleteDone * if !&previewwindow && &completeopt =~ 'preview' | silent! pclose | endif
  endif
augroup END

augroup go
  au!
  au Filetype go command! -bang A call go#alternate#Switch(<bang>0, 'edit')
  au Filetype go command! -bang AV call go#alternate#Switch(<bang>0, 'vsplit')
  au Filetype go command! -bang AS call go#alternate#Switch(<bang>0, 'split')
  au Filetype go command! -bang AT call go#alternate#Switch(<bang>0, 'tabe')

  au FileType go nmap <Leader>dd <Plug>(go-def-vertical)
  au FileType go nmap <Leader>dv <Plug>(go-doc-vertical)
  au FileType go nmap <Leader>db <Plug>(go-doc-browser)

  au FileType go nmap <leader>r  <Plug>(go-run)
  au FileType go nmap <leader>t  <Plug>(go-test)
  au FileType go nmap <Leader>gt <Plug>(go-coverage-toggle)
  au FileType go nmap <Leader>i <Plug>(go-info)
  au FileType go nmap <silent> <Leader>l <Plug>(go-metalinter)
  au FileType go nmap <C-g> :GoDecls<cr>
  au FileType go nmap <leader>dr :GoDeclsDir<cr>
  au FileType go imap <C-g> <esc>:<C-u>GoDecls<cr>
  au FileType go imap <leader>dr <esc>:<C-u>GoDeclsDir<cr>
  au FileType go nmap <leader>rb :<C-u>call <SID>build_go_files()<CR>
augroup END

" vim-airline configuration
if !exists('g:airline_symbols')
  let g:airline_symbols = {}
endif

if !exists('g:airline_powerline_fonts')
  let g:airline#extensions#tabline#left_sep = ' '
  let g:airline#extensions#tabline#left_alt_sep = '|'
  let g:airline_left_sep = '▶'
  let g:airline_left_alt_sep = '»'
  let g:airline_right_sep = '◀'
  let g:airline_right_alt_sep = '«'
  let g:airline#extensions#branch#prefix = '⤴' "➔, ➥, ⎇
  let g:airline#extensions#readonly#symbol = '⊘'
  let g:airline#extensions#linecolumn#prefix = '¶'
  let g:airline#extensions#paste#symbol = 'ρ'
  let g:airline_symbols.linenr = '␊'
  let g:airline_symbols.branch = '⎇'
  let g:airline_symbols.paste = 'ρ'
  let g:airline_symbols.paste = 'Þ'
  let g:airline_symbols.paste = '∥'
  let g:airline_symbols.whitespace = 'Ξ'
else
  let g:airline#extensions#tabline#left_sep = ''
  let g:airline#extensions#tabline#left_alt_sep = ''

  " Powerline symbols
  let g:airline_left_sep = ''
  let g:airline_left_alt_sep = ''
  let g:airline_right_sep = ''
  let g:airline_right_alt_sep = ''
  let g:airline_symbols.branch = ''
  let g:airline_symbols.readonly = ''
  let g:airline_symbols.linenr = ''
endif

" NERDTree configuration
autocmd vimenter * if !argc() | NERDTree | endif
autocmd bufenter * if (winnr('$') == 1 && exists('b:NERDTree') && b:NERDTree.isTabTree()) | q | endif
autocmd StdinReadPre * let s:std_in=1
autocmd VimEnter * if argc() > 0 && !exists('s:std_in') && argc() == 1 && isdirectory(argv()[0]) | exe 'NERDTree ' . argv()[0] | wincmd p | ene | endif
nmap <C-n> :NERDTreeToggle<CR>
nmap <leader>n :NERDTreeFind<CR>
let NERDTreeIgnore=['\.pyc$', '\~$', '\.swp$']
let NERDTreeShowLineNumbers=1
let g:NERDTreeMapOpenSplit = 's'
let g:NERDTreeMapOpenVSplit = 'v'
autocmd BufEnter * if (winnr('$') == 1 && &filetype == 'nerdtree') | q | endif
let g:NERDTreeWinSize=31

" Set the colorscheme to material
colorscheme material

" Enable 24-bit RGB colors in the TUI
if (has("termguicolors"))
  set termguicolors
endif

" Set background to dark for better visibility
set background=dark
