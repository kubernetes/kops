/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"io"

	"bytes"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

const boilerPlate = `
# Copyright 2019 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
`

var (
	completionShells = map[string]func(out io.Writer, cmd *cobra.Command) error{
		"bash": runCompletionBash,
		"zsh":  runCompletionZsh,
	}
)

type CompletionOptions struct {
	Shell string
}

var (
	completion_long = templates.LongDesc(i18n.T(`
	Output shell completion code for the specified shell (bash or zsh).
	The shell code must be evaluated to provide interactive
	completion of kops commands.  This can be done by sourcing it from
	the .bash_profile.

	Note: this requires the bash-completion framework, which is not installed
	by default on Mac. Once installed, bash_completion must be evaluated.  This can be done by adding the
	following line to the .bash_profile


	Note for zsh users: zsh completions are only supported in versions of zsh >= 5.2`))

	completion_example = templates.Examples(i18n.T(`
	# For OSX users install bash completion using homebrew
	brew install bash-completion
	source $(brew --prefix)/etc/bash_completion

	# Bash completion support
	printf "source $(brew --prefix)/etc/bash_completion\n" >> $HOME/.bash_profile
	source $HOME/.bash_profile
	source <(kops completion bash)
	kops completion bash > ~/.kops/completion.bash.inc
	chmod +x $HOME/.kops/completion.bash.inc

	# kops shell completion
	printf "$HOME/.kops/completion.bash.inc\n" >> $HOME/.bash_profile
	source $HOME/.bash_profile

	# Load the kops completion code for zsh[1] into the current shell
	source <(kops completion zsh)`))

	completion_short = i18n.T("Output shell completion code for the given shell (bash or zsh).")
)

func NewCmdCompletion(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CompletionOptions{}

	cmd := &cobra.Command{
		Use:     "completion",
		Short:   completion_short,
		Long:    completion_long,
		Example: completion_example,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunCompletion(f, cmd, args, out, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().StringVar(&options.Shell, "shell", "", "target shell (bash).")

	return cmd
}

func RunCompletion(f *util.Factory, cmd *cobra.Command, args []string, out io.Writer, c *CompletionOptions) error {
	if len(args) != 0 {

		if c.Shell != "" {
			return fmt.Errorf("cannot specify shell both as a flag and a positional argument")
		}
		c.Shell = args[0]
	}

	if c.Shell == "" {
		return fmt.Errorf("shell is required")
	}

	run, found := completionShells[c.Shell]
	if !found {
		return fmt.Errorf("Unsupported shell type %q.", args[0])
	}

	if _, err := out.Write([]byte(boilerPlate)); err != nil {
		return err
	}

	return run(out, cmd.Parent())
}

func runCompletionBash(out io.Writer, cmd *cobra.Command) error {
	return cmd.GenBashCompletion(out)
}

func runCompletionZsh(out io.Writer, cmd *cobra.Command) error {
	zsh_head := "#compdef kops\n"

	out.Write([]byte(zsh_head))

	zsh_initialization := `
__kops_bash_source() {
	alias shopt=':'
	alias _expand=_bash_expand
	alias _complete=_bash_comp
	emulate -L sh
	setopt kshglob noshglob braceexpand
	source "$@"
}
__kops_type() {
	# -t is not supported by zsh
	if [ "$1" == "-t" ]; then
		shift
		# fake Bash 4 to disable "complete -o nospace". Instead
		# "compopt +-o nospace" is used in the code to toggle trailing
		# spaces. We don't support that, but leave trailing spaces on
		# all the time
		if [ "$1" = "__kops_compopt" ]; then
			echo builtin
			return 0
		fi
	fi
	type "$@"
}
__kops_compgen() {
	local completions w
	completions=( $(compgen "$@") ) || return $?
	# filter by given word as prefix
	while [[ "$1" = -* && "$1" != -- ]]; do
		shift
		shift
	done
	if [[ "$1" == -- ]]; then
		shift
	fi
	for w in "${completions[@]}"; do
		if [[ "${w}" = "$1"* ]]; then
			echo "${w}"
		fi
	done
}
__kops_compopt() {
	true # don't do anything. Not supported by bashcompinit in zsh
}

__kops_ltrim_colon_completions()
{
	if [[ "$1" == *:* && "$COMP_WORDBREAKS" == *:* ]]; then
		# Remove colon-word prefix from COMPREPLY items
		local colon_word=${1%${1##*:}}
		local i=${#COMPREPLY[*]}
		while [[ $((--i)) -ge 0 ]]; do
			COMPREPLY[$i]=${COMPREPLY[$i]#"$colon_word"}
		done
	fi
}
__kops_get_comp_words_by_ref() {
	cur="${COMP_WORDS[COMP_CWORD]}"
	prev="${COMP_WORDS[${COMP_CWORD}-1]}"
	words=("${COMP_WORDS[@]}")
	cword=("${COMP_CWORD[@]}")
}
__kops_filedir() {
	local RET OLD_IFS w qw
	__debug "_filedir $@ cur=$cur"
	if [[ "$1" = \~* ]]; then
		# somehow does not work. Maybe, zsh does not call this at all
		eval echo "$1"
		return 0
	fi
	OLD_IFS="$IFS"
	IFS=$'\n'
	if [ "$1" = "-d" ]; then
		shift
		RET=( $(compgen -d) )
	else
		RET=( $(compgen -f) )
	fi
	IFS="$OLD_IFS"
	IFS="," __debug "RET=${RET[@]} len=${#RET[@]}"
	for w in ${RET[@]}; do
		if [[ ! "${w}" = "${cur}"* ]]; then
			continue
		fi
		if eval "[[ \"\${w}\" = *.$1 || -d \"\${w}\" ]]"; then
			qw="$(__kops_quote "${w}")"
			if [ -d "${w}" ]; then
				COMPREPLY+=("${qw}/")
			else
				COMPREPLY+=("${qw}")
			fi
		fi
	done
}
__kops_quote() {
    if [[ $1 == \'* || $1 == \"* ]]; then
        # Leave out first character
        printf %q "${1:1}"
    else
    	printf %q "$1"
    fi
}
autoload -U +X bashcompinit && bashcompinit
# use word boundary patterns for BSD or GNU sed
LWORD='[[:<:]]'
RWORD='[[:>:]]'
if sed --help 2>&1 | grep -q GNU; then
	LWORD='\<'
	RWORD='\>'
fi
__kops_convert_bash_to_zsh() {
	sed \
	-e 's/declare -F/whence -w/' \
	-e 's/_get_comp_words_by_ref "\$@"/_get_comp_words_by_ref "\$*"/' \
	-e 's/local \([a-zA-Z0-9_]*\)=/local \1; \1=/' \
	-e 's/flags+=("\(--.*\)=")/flags+=("\1"); two_word_flags+=("\1")/' \
	-e 's/must_have_one_flag+=("\(--.*\)=")/must_have_one_flag+=("\1")/' \
	-e "s/${LWORD}_filedir${RWORD}/__kops_filedir/g" \
	-e "s/${LWORD}_get_comp_words_by_ref${RWORD}/__kops_get_comp_words_by_ref/g" \
	-e "s/${LWORD}__ltrim_colon_completions${RWORD}/__kops_ltrim_colon_completions/g" \
	-e "s/${LWORD}compgen${RWORD}/__kops_compgen/g" \
	-e "s/${LWORD}compopt${RWORD}/__kops_compopt/g" \
	-e "s/${LWORD}declare${RWORD}/builtin declare/g" \
	-e "s/\\\$(type${RWORD}/\$(__kops_type/g" \
	<<'BASH_COMPLETION_EOF'
`
	out.Write([]byte(zsh_initialization))

	buf := new(bytes.Buffer)
	cmd.GenBashCompletion(buf)
	out.Write(buf.Bytes())

	zsh_tail := `
BASH_COMPLETION_EOF
}
__kops_bash_source <(__kops_convert_bash_to_zsh)
_complete kops 2>/dev/null
`
	out.Write([]byte(zsh_tail))
	return nil
}
