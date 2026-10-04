# bash completion for gocat                                -*- shell-script -*-

__gocat_debug()
{
    if [[ -n ${BASH_COMP_DEBUG_FILE:-} ]]; then
        echo "$*" >> "${BASH_COMP_DEBUG_FILE}"
    fi
}

# Homebrew on Macs have version 1.3 of bash-completion which doesn't include
# _init_completion. This is a very minimal version of that function.
__gocat_init_completion()
{
    COMPREPLY=()
    _get_comp_words_by_ref "$@" cur prev words cword
}

__gocat_index_of_word()
{
    local w word=$1
    shift
    index=0
    for w in "$@"; do
        [[ $w = "$word" ]] && return
        index=$((index+1))
    done
    index=-1
}

__gocat_contains_word()
{
    local w word=$1; shift
    for w in "$@"; do
        [[ $w = "$word" ]] && return
    done
    return 1
}

__gocat_handle_go_custom_completion()
{
    __gocat_debug "${FUNCNAME[0]}: cur is ${cur}, words[*] is ${words[*]}, #words[@] is ${#words[@]}"

    local shellCompDirectiveError=1
    local shellCompDirectiveNoSpace=2
    local shellCompDirectiveNoFileComp=4
    local shellCompDirectiveFilterFileExt=8
    local shellCompDirectiveFilterDirs=16

    local out requestComp lastParam lastChar comp directive args

    # Prepare the command to request completions for the program.
    # Calling ${words[0]} instead of directly gocat allows handling aliases
    args=("${words[@]:1}")
    # Disable ActiveHelp which is not supported for bash completion v1
    requestComp="GOCAT_ACTIVE_HELP=0 ${words[0]} __completeNoDesc ${args[*]}"

    lastParam=${words[$((${#words[@]}-1))]}
    lastChar=${lastParam:$((${#lastParam}-1)):1}
    __gocat_debug "${FUNCNAME[0]}: lastParam ${lastParam}, lastChar ${lastChar}"

    if [ -z "${cur}" ] && [ "${lastChar}" != "=" ]; then
        # If the last parameter is complete (there is a space following it)
        # We add an extra empty parameter so we can indicate this to the go method.
        __gocat_debug "${FUNCNAME[0]}: Adding extra empty parameter"
        requestComp="${requestComp} \"\""
    fi

    __gocat_debug "${FUNCNAME[0]}: calling ${requestComp}"
    # Use eval to handle any environment variables and such
    out=$(eval "${requestComp}" 2>/dev/null)

    # Extract the directive integer at the very end of the output following a colon (:)
    directive=${out##*:}
    # Remove the directive
    out=${out%:*}
    if [ "${directive}" = "${out}" ]; then
        # There is not directive specified
        directive=0
    fi
    __gocat_debug "${FUNCNAME[0]}: the completion directive is: ${directive}"
    __gocat_debug "${FUNCNAME[0]}: the completions are: ${out}"

    if [ $((directive & shellCompDirectiveError)) -ne 0 ]; then
        # Error code.  No completion.
        __gocat_debug "${FUNCNAME[0]}: received error from custom completion go code"
        return
    else
        if [ $((directive & shellCompDirectiveNoSpace)) -ne 0 ]; then
            if [[ $(type -t compopt) = "builtin" ]]; then
                __gocat_debug "${FUNCNAME[0]}: activating no space"
                compopt -o nospace
            fi
        fi
        if [ $((directive & shellCompDirectiveNoFileComp)) -ne 0 ]; then
            if [[ $(type -t compopt) = "builtin" ]]; then
                __gocat_debug "${FUNCNAME[0]}: activating no file completion"
                compopt +o default
            fi
        fi
    fi

    if [ $((directive & shellCompDirectiveFilterFileExt)) -ne 0 ]; then
        # File extension filtering
        local fullFilter filter filteringCmd
        # Do not use quotes around the $out variable or else newline
        # characters will be kept.
        for filter in ${out}; do
            fullFilter+="$filter|"
        done

        filteringCmd="_filedir $fullFilter"
        __gocat_debug "File filtering command: $filteringCmd"
        $filteringCmd
    elif [ $((directive & shellCompDirectiveFilterDirs)) -ne 0 ]; then
        # File completion for directories only
        local subdir
        # Use printf to strip any trailing newline
        subdir=$(printf "%s" "${out}")
        if [ -n "$subdir" ]; then
            __gocat_debug "Listing directories in $subdir"
            __gocat_handle_subdirs_in_dir_flag "$subdir"
        else
            __gocat_debug "Listing directories in ."
            _filedir -d
        fi
    else
        while IFS='' read -r comp; do
            COMPREPLY+=("$comp")
        done < <(compgen -W "${out}" -- "$cur")
    fi
}

__gocat_handle_reply()
{
    __gocat_debug "${FUNCNAME[0]}"
    local comp
    case $cur in
        -*)
            if [[ $(type -t compopt) = "builtin" ]]; then
                compopt -o nospace
            fi
            local allflags
            if [ ${#must_have_one_flag[@]} -ne 0 ]; then
                allflags=("${must_have_one_flag[@]}")
            else
                allflags=("${flags[*]} ${two_word_flags[*]}")
            fi
            while IFS='' read -r comp; do
                COMPREPLY+=("$comp")
            done < <(compgen -W "${allflags[*]}" -- "$cur")
            if [[ $(type -t compopt) = "builtin" ]]; then
                [[ "${COMPREPLY[0]}" == *= ]] || compopt +o nospace
            fi

            # complete after --flag=abc
            if [[ $cur == *=* ]]; then
                if [[ $(type -t compopt) = "builtin" ]]; then
                    compopt +o nospace
                fi

                local index flag
                flag="${cur%=*}"
                __gocat_index_of_word "${flag}" "${flags_with_completion[@]}"
                COMPREPLY=()
                if [[ ${index} -ge 0 ]]; then
                    PREFIX=""
                    cur="${cur#*=}"
                    ${flags_completion[${index}]}
                    if [ -n "${ZSH_VERSION:-}" ]; then
                        # zsh completion needs --flag= prefix
                        eval "COMPREPLY=( \"\${COMPREPLY[@]/#/${flag}=}\" )"
                    fi
                fi
            fi

            if [[ -z "${flag_parsing_disabled}" ]]; then
                # If flag parsing is enabled, we have completed the flags and can return.
                # If flag parsing is disabled, we may not know all (or any) of the flags, so we fallthrough
                # to possibly call handle_go_custom_completion.
                return 0;
            fi
            ;;
    esac

    # check if we are handling a flag with special work handling
    local index
    __gocat_index_of_word "${prev}" "${flags_with_completion[@]}"
    if [[ ${index} -ge 0 ]]; then
        ${flags_completion[${index}]}
        return
    fi

    # we are parsing a flag and don't have a special handler, no completion
    if [[ ${cur} != "${words[cword]}" ]]; then
        return
    fi

    local completions
    completions=("${commands[@]}")
    if [[ ${#must_have_one_noun[@]} -ne 0 ]]; then
        completions+=("${must_have_one_noun[@]}")
    elif [[ -n "${has_completion_function}" ]]; then
        # if a go completion function is provided, defer to that function
        __gocat_handle_go_custom_completion
    fi
    if [[ ${#must_have_one_flag[@]} -ne 0 ]]; then
        completions+=("${must_have_one_flag[@]}")
    fi
    while IFS='' read -r comp; do
        COMPREPLY+=("$comp")
    done < <(compgen -W "${completions[*]}" -- "$cur")

    if [[ ${#COMPREPLY[@]} -eq 0 && ${#noun_aliases[@]} -gt 0 && ${#must_have_one_noun[@]} -ne 0 ]]; then
        while IFS='' read -r comp; do
            COMPREPLY+=("$comp")
        done < <(compgen -W "${noun_aliases[*]}" -- "$cur")
    fi

    if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
        if declare -F __gocat_custom_func >/dev/null; then
            # try command name qualified custom func
            __gocat_custom_func
        else
            # otherwise fall back to unqualified for compatibility
            declare -F __custom_func >/dev/null && __custom_func
        fi
    fi

    # available in bash-completion >= 2, not always present on macOS
    if declare -F __ltrim_colon_completions >/dev/null; then
        __ltrim_colon_completions "$cur"
    fi

    # If there is only 1 completion and it is a flag with an = it will be completed
    # but we don't want a space after the =
    if [[ "${#COMPREPLY[@]}" -eq "1" ]] && [[ $(type -t compopt) = "builtin" ]] && [[ "${COMPREPLY[0]}" == --*= ]]; then
       compopt -o nospace
    fi
}

# The arguments should be in the form "ext1|ext2|extn"
__gocat_handle_filename_extension_flag()
{
    local ext="$1"
    _filedir "@(${ext})"
}

__gocat_handle_subdirs_in_dir_flag()
{
    local dir="$1"
    pushd "${dir}" >/dev/null 2>&1 && _filedir -d && popd >/dev/null 2>&1 || return
}

__gocat_handle_flag()
{
    __gocat_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    # if a command required a flag, and we found it, unset must_have_one_flag()
    local flagname=${words[c]}
    local flagvalue=""
    # if the word contained an =
    if [[ ${words[c]} == *"="* ]]; then
        flagvalue=${flagname#*=} # take in as flagvalue after the =
        flagname=${flagname%=*} # strip everything after the =
        flagname="${flagname}=" # but put the = back
    fi
    __gocat_debug "${FUNCNAME[0]}: looking for ${flagname}"
    if __gocat_contains_word "${flagname}" "${must_have_one_flag[@]}"; then
        must_have_one_flag=()
    fi

    # if you set a flag which only applies to this command, don't show subcommands
    if __gocat_contains_word "${flagname}" "${local_nonpersistent_flags[@]}"; then
      commands=()
    fi

    # keep flag value with flagname as flaghash
    # flaghash variable is an associative array which is only supported in bash > 3.
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        if [ -n "${flagvalue}" ] ; then
            flaghash[${flagname}]=${flagvalue}
        elif [ -n "${words[ $((c+1)) ]}" ] ; then
            flaghash[${flagname}]=${words[ $((c+1)) ]}
        else
            flaghash[${flagname}]="true" # pad "true" for bool flag
        fi
    fi

    # skip the argument to a two word flag
    if [[ ${words[c]} != *"="* ]] && __gocat_contains_word "${words[c]}" "${two_word_flags[@]}"; then
        __gocat_debug "${FUNCNAME[0]}: found a flag ${words[c]}, skip the next argument"
        c=$((c+1))
        # if we are looking for a flags value, don't show commands
        if [[ $c -eq $cword ]]; then
            commands=()
        fi
    fi

    c=$((c+1))

}

__gocat_handle_noun()
{
    __gocat_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    if __gocat_contains_word "${words[c]}" "${must_have_one_noun[@]}"; then
        must_have_one_noun=()
    elif __gocat_contains_word "${words[c]}" "${noun_aliases[@]}"; then
        must_have_one_noun=()
    fi

    nouns+=("${words[c]}")
    c=$((c+1))
}

__gocat_handle_command()
{
    __gocat_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    local next_command
    if [[ -n ${last_command} ]]; then
        next_command="_${last_command}_${words[c]//:/__}"
    else
        if [[ $c -eq 0 ]]; then
            next_command="_gocat_root_command"
        else
            next_command="_${words[c]//:/__}"
        fi
    fi
    c=$((c+1))
    __gocat_debug "${FUNCNAME[0]}: looking for ${next_command}"
    declare -F "$next_command" >/dev/null && $next_command
}

__gocat_handle_word()
{
    if [[ $c -ge $cword ]]; then
        __gocat_handle_reply
        return
    fi
    __gocat_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"
    if [[ "${words[c]}" == -* ]]; then
        __gocat_handle_flag
    elif __gocat_contains_word "${words[c]}" "${commands[@]}"; then
        __gocat_handle_command
    elif [[ $c -eq 0 ]]; then
        __gocat_handle_command
    elif __gocat_contains_word "${words[c]}" "${command_aliases[@]}"; then
        # aliashash variable is an associative array which is only supported in bash > 3.
        if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
            words[c]=${aliashash[${words[c]}]}
            __gocat_handle_command
        else
            __gocat_handle_noun
        fi
    else
        __gocat_handle_noun
    fi
    __gocat_handle_word
}

_gocat_benchmark()
{
    last_command="gocat_benchmark"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--bench-verbose")
    local_nonpersistent_flags+=("--bench-verbose")
    flags+=("--connections=")
    two_word_flags+=("--connections")
    local_nonpersistent_flags+=("--connections")
    local_nonpersistent_flags+=("--connections=")
    flags+=("--duration=")
    two_word_flags+=("--duration")
    local_nonpersistent_flags+=("--duration")
    local_nonpersistent_flags+=("--duration=")
    flags+=("--packet-size=")
    two_word_flags+=("--packet-size")
    local_nonpersistent_flags+=("--packet-size")
    local_nonpersistent_flags+=("--packet-size=")
    flags+=("--port=")
    two_word_flags+=("--port")
    local_nonpersistent_flags+=("--port")
    local_nonpersistent_flags+=("--port=")
    flags+=("--protocol=")
    two_word_flags+=("--protocol")
    local_nonpersistent_flags+=("--protocol")
    local_nonpersistent_flags+=("--protocol=")
    flags+=("--rate=")
    two_word_flags+=("--rate")
    local_nonpersistent_flags+=("--rate")
    local_nonpersistent_flags+=("--rate=")
    flags+=("--target=")
    two_word_flags+=("--target")
    local_nonpersistent_flags+=("--target")
    local_nonpersistent_flags+=("--target=")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_flag+=("--target=")
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_broker()
{
    last_command="gocat_broker"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--max-conns=")
    two_word_flags+=("--max-conns")
    two_word_flags+=("-m")
    local_nonpersistent_flags+=("--max-conns")
    local_nonpersistent_flags+=("--max-conns=")
    local_nonpersistent_flags+=("-m")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_chat()
{
    last_command="gocat_chat"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--max-conns=")
    two_word_flags+=("--max-conns")
    two_word_flags+=("-m")
    local_nonpersistent_flags+=("--max-conns")
    local_nonpersistent_flags+=("--max-conns=")
    local_nonpersistent_flags+=("-m")
    flags+=("--room=")
    two_word_flags+=("--room")
    two_word_flags+=("-r")
    local_nonpersistent_flags+=("--room")
    local_nonpersistent_flags+=("--room=")
    local_nonpersistent_flags+=("-r")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_completion_bash()
{
    last_command="gocat_completion_bash"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--help")
    flags+=("-h")
    local_nonpersistent_flags+=("--help")
    local_nonpersistent_flags+=("-h")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_completion_fish()
{
    last_command="gocat_completion_fish"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_completion_install()
{
    last_command="gocat_completion_install"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_completion_powershell()
{
    last_command="gocat_completion_powershell"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_completion_zsh()
{
    last_command="gocat_completion_zsh"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_completion()
{
    last_command="gocat_completion"

    command_aliases=()

    commands=()
    commands+=("bash")
    commands+=("fish")
    commands+=("install")
    commands+=("powershell")
    commands+=("zsh")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_connect()
{
    last_command="gocat_connect"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--auto-reconnect")
    local_nonpersistent_flags+=("--auto-reconnect")
    flags+=("--backoff=")
    two_word_flags+=("--backoff")
    local_nonpersistent_flags+=("--backoff")
    local_nonpersistent_flags+=("--backoff=")
    flags+=("--backoff-base=")
    two_word_flags+=("--backoff-base")
    local_nonpersistent_flags+=("--backoff-base")
    local_nonpersistent_flags+=("--backoff-base=")
    flags+=("--backoff-jitter")
    local_nonpersistent_flags+=("--backoff-jitter")
    flags+=("--backoff-max=")
    two_word_flags+=("--backoff-max")
    local_nonpersistent_flags+=("--backoff-max")
    local_nonpersistent_flags+=("--backoff-max=")
    flags+=("--ca-cert=")
    two_word_flags+=("--ca-cert")
    local_nonpersistent_flags+=("--ca-cert")
    local_nonpersistent_flags+=("--ca-cert=")
    flags+=("--connect-timeout=")
    two_word_flags+=("--connect-timeout")
    local_nonpersistent_flags+=("--connect-timeout")
    local_nonpersistent_flags+=("--connect-timeout=")
    flags+=("--keep-alive")
    local_nonpersistent_flags+=("--keep-alive")
    flags+=("--proxy=")
    two_word_flags+=("--proxy")
    local_nonpersistent_flags+=("--proxy")
    local_nonpersistent_flags+=("--proxy=")
    flags+=("--retry=")
    two_word_flags+=("--retry")
    local_nonpersistent_flags+=("--retry")
    local_nonpersistent_flags+=("--retry=")
    flags+=("--shell=")
    two_word_flags+=("--shell")
    local_nonpersistent_flags+=("--shell")
    local_nonpersistent_flags+=("--shell=")
    flags+=("--udp")
    local_nonpersistent_flags+=("--udp")
    flags+=("--verify-cert")
    local_nonpersistent_flags+=("--verify-cert")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_console()
{
    last_command="gocat_console"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_convert()
{
    last_command="gocat_convert"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--buffer=")
    two_word_flags+=("--buffer")
    local_nonpersistent_flags+=("--buffer")
    local_nonpersistent_flags+=("--buffer=")
    flags+=("--from=")
    two_word_flags+=("--from")
    local_nonpersistent_flags+=("--from")
    local_nonpersistent_flags+=("--from=")
    flags+=("--to=")
    two_word_flags+=("--to")
    local_nonpersistent_flags+=("--to")
    local_nonpersistent_flags+=("--to=")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_flag+=("--from=")
    must_have_one_flag+=("--to=")
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_distributed()
{
    last_command="gocat_distributed"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--id=")
    two_word_flags+=("--id")
    local_nonpersistent_flags+=("--id")
    local_nonpersistent_flags+=("--id=")
    flags+=("--master=")
    two_word_flags+=("--master")
    local_nonpersistent_flags+=("--master")
    local_nonpersistent_flags+=("--master=")
    flags+=("--max-workers=")
    two_word_flags+=("--max-workers")
    local_nonpersistent_flags+=("--max-workers")
    local_nonpersistent_flags+=("--max-workers=")
    flags+=("--mode=")
    two_word_flags+=("--mode")
    local_nonpersistent_flags+=("--mode")
    local_nonpersistent_flags+=("--mode=")
    flags+=("--port=")
    two_word_flags+=("--port")
    local_nonpersistent_flags+=("--port")
    local_nonpersistent_flags+=("--port=")
    flags+=("--token=")
    two_word_flags+=("--token")
    local_nonpersistent_flags+=("--token")
    local_nonpersistent_flags+=("--token=")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_dns-tunnel()
{
    last_command="gocat_dns-tunnel"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--client")
    local_nonpersistent_flags+=("--client")
    flags+=("--dns-port=")
    two_word_flags+=("--dns-port")
    local_nonpersistent_flags+=("--dns-port")
    local_nonpersistent_flags+=("--dns-port=")
    flags+=("--dns-server=")
    two_word_flags+=("--dns-server")
    local_nonpersistent_flags+=("--dns-server")
    local_nonpersistent_flags+=("--dns-server=")
    flags+=("--domain=")
    two_word_flags+=("--domain")
    local_nonpersistent_flags+=("--domain")
    local_nonpersistent_flags+=("--domain=")
    flags+=("--encoding=")
    two_word_flags+=("--encoding")
    local_nonpersistent_flags+=("--encoding")
    local_nonpersistent_flags+=("--encoding=")
    flags+=("--listen=")
    two_word_flags+=("--listen")
    local_nonpersistent_flags+=("--listen")
    local_nonpersistent_flags+=("--listen=")
    flags+=("--server")
    local_nonpersistent_flags+=("--server")
    flags+=("--target=")
    two_word_flags+=("--target")
    local_nonpersistent_flags+=("--target")
    local_nonpersistent_flags+=("--target=")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_flag+=("--domain=")
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_doctor()
{
    last_command="gocat_doctor"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--commands")
    local_nonpersistent_flags+=("--commands")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_help()
{
    last_command="gocat_help"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    has_completion_function=1
    noun_aliases=()
}

_gocat_interfaces()
{
    last_command="gocat_interfaces"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_listen()
{
    last_command="gocat_listen"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--auto-upgrade")
    local_nonpersistent_flags+=("--auto-upgrade")
    flags+=("--bind=")
    two_word_flags+=("--bind")
    local_nonpersistent_flags+=("--bind")
    local_nonpersistent_flags+=("--bind=")
    flags+=("--block-signals")
    local_nonpersistent_flags+=("--block-signals")
    flags+=("--iface=")
    two_word_flags+=("--iface")
    local_nonpersistent_flags+=("--iface")
    local_nonpersistent_flags+=("--iface=")
    flags+=("--interactive")
    local_nonpersistent_flags+=("--interactive")
    flags+=("--listen-exec=")
    two_word_flags+=("--listen-exec")
    local_nonpersistent_flags+=("--listen-exec")
    local_nonpersistent_flags+=("--listen-exec=")
    flags+=("--listen-ipv4")
    local_nonpersistent_flags+=("--listen-ipv4")
    flags+=("--listen-ipv6")
    local_nonpersistent_flags+=("--listen-ipv6")
    flags+=("--listen-keep-alive")
    local_nonpersistent_flags+=("--listen-keep-alive")
    flags+=("--listen-max-conn=")
    two_word_flags+=("--listen-max-conn")
    local_nonpersistent_flags+=("--listen-max-conn")
    local_nonpersistent_flags+=("--listen-max-conn=")
    flags+=("--listen-ssl")
    local_nonpersistent_flags+=("--listen-ssl")
    flags+=("--listen-ssl-cert=")
    two_word_flags+=("--listen-ssl-cert")
    local_nonpersistent_flags+=("--listen-ssl-cert")
    local_nonpersistent_flags+=("--listen-ssl-cert=")
    flags+=("--listen-ssl-key=")
    two_word_flags+=("--listen-ssl-key")
    local_nonpersistent_flags+=("--listen-ssl-key")
    local_nonpersistent_flags+=("--listen-ssl-key=")
    flags+=("--listen-timeout=")
    two_word_flags+=("--listen-timeout")
    local_nonpersistent_flags+=("--listen-timeout")
    local_nonpersistent_flags+=("--listen-timeout=")
    flags+=("--listen-udp")
    local_nonpersistent_flags+=("--listen-udp")
    flags+=("--local")
    local_nonpersistent_flags+=("--local")
    flags+=("--maintain=")
    two_word_flags+=("--maintain")
    local_nonpersistent_flags+=("--maintain")
    local_nonpersistent_flags+=("--maintain=")
    flags+=("--max-sessions=")
    two_word_flags+=("--max-sessions")
    local_nonpersistent_flags+=("--max-sessions")
    local_nonpersistent_flags+=("--max-sessions=")
    flags+=("--no-attach")
    local_nonpersistent_flags+=("--no-attach")
    flags+=("--payloads")
    local_nonpersistent_flags+=("--payloads")
    flags+=("--session")
    local_nonpersistent_flags+=("--session")
    flags+=("--single-session")
    local_nonpersistent_flags+=("--single-session")
    flags+=("--ssh-trigger=")
    two_word_flags+=("--ssh-trigger")
    local_nonpersistent_flags+=("--ssh-trigger")
    local_nonpersistent_flags+=("--ssh-trigger=")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_mcp_setup()
{
    last_command="gocat_mcp_setup"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--client=")
    two_word_flags+=("--client")
    local_nonpersistent_flags+=("--client")
    local_nonpersistent_flags+=("--client=")
    flags+=("--list")
    local_nonpersistent_flags+=("--list")
    flags+=("--remove")
    local_nonpersistent_flags+=("--remove")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_mcp()
{
    last_command="gocat_mcp"

    command_aliases=()

    commands=()
    commands+=("setup")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--name=")
    two_word_flags+=("--name")
    local_nonpersistent_flags+=("--name")
    local_nonpersistent_flags+=("--name=")
    flags+=("--version=")
    two_word_flags+=("--version")
    local_nonpersistent_flags+=("--version")
    local_nonpersistent_flags+=("--version=")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_metrics()
{
    last_command="gocat_metrics"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--interval=")
    two_word_flags+=("--interval")
    local_nonpersistent_flags+=("--interval")
    local_nonpersistent_flags+=("--interval=")
    flags+=("--namespace=")
    two_word_flags+=("--namespace")
    local_nonpersistent_flags+=("--namespace")
    local_nonpersistent_flags+=("--namespace=")
    flags+=("--port=")
    two_word_flags+=("--port")
    local_nonpersistent_flags+=("--port")
    local_nonpersistent_flags+=("--port=")
    flags+=("--subsystem=")
    two_word_flags+=("--subsystem")
    local_nonpersistent_flags+=("--subsystem")
    local_nonpersistent_flags+=("--subsystem=")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_multi-listen()
{
    last_command="gocat_multi-listen"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--bind=")
    two_word_flags+=("--bind")
    local_nonpersistent_flags+=("--bind")
    local_nonpersistent_flags+=("--bind=")
    flags+=("--exec=")
    two_word_flags+=("--exec")
    local_nonpersistent_flags+=("--exec")
    local_nonpersistent_flags+=("--exec=")
    flags+=("--max-connections=")
    two_word_flags+=("--max-connections")
    local_nonpersistent_flags+=("--max-connections")
    local_nonpersistent_flags+=("--max-connections=")
    flags+=("--ports=")
    two_word_flags+=("--ports")
    local_nonpersistent_flags+=("--ports")
    local_nonpersistent_flags+=("--ports=")
    flags+=("--range=")
    two_word_flags+=("--range")
    local_nonpersistent_flags+=("--range")
    local_nonpersistent_flags+=("--range=")
    flags+=("--stats")
    local_nonpersistent_flags+=("--stats")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_payload()
{
    last_command="gocat_payload"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--all")
    flags+=("-a")
    local_nonpersistent_flags+=("--all")
    local_nonpersistent_flags+=("-a")
    flags+=("--encode")
    flags+=("-E")
    local_nonpersistent_flags+=("--encode")
    local_nonpersistent_flags+=("-E")
    flags+=("--format=")
    two_word_flags+=("--format")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--format")
    local_nonpersistent_flags+=("--format=")
    local_nonpersistent_flags+=("-f")
    flags+=("--interface=")
    two_word_flags+=("--interface")
    two_word_flags+=("-I")
    local_nonpersistent_flags+=("--interface")
    local_nonpersistent_flags+=("--interface=")
    local_nonpersistent_flags+=("-I")
    flags+=("--type=")
    two_word_flags+=("--type")
    two_word_flags+=("-T")
    local_nonpersistent_flags+=("--type")
    local_nonpersistent_flags+=("--type=")
    local_nonpersistent_flags+=("-T")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_portforward()
{
    last_command="gocat_portforward"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--buffer-size=")
    two_word_flags+=("--buffer-size")
    local_nonpersistent_flags+=("--buffer-size")
    local_nonpersistent_flags+=("--buffer-size=")
    flags+=("--idle-timeout=")
    two_word_flags+=("--idle-timeout")
    local_nonpersistent_flags+=("--idle-timeout")
    local_nonpersistent_flags+=("--idle-timeout=")
    flags+=("--local-port=")
    two_word_flags+=("--local-port")
    local_nonpersistent_flags+=("--local-port")
    local_nonpersistent_flags+=("--local-port=")
    flags+=("--max-conns=")
    two_word_flags+=("--max-conns")
    local_nonpersistent_flags+=("--max-conns")
    local_nonpersistent_flags+=("--max-conns=")
    flags+=("--pf-verbose")
    local_nonpersistent_flags+=("--pf-verbose")
    flags+=("--protocol=")
    two_word_flags+=("--protocol")
    local_nonpersistent_flags+=("--protocol")
    local_nonpersistent_flags+=("--protocol=")
    flags+=("--remote=")
    two_word_flags+=("--remote")
    two_word_flags+=("-r")
    local_nonpersistent_flags+=("--remote")
    local_nonpersistent_flags+=("--remote=")
    local_nonpersistent_flags+=("-r")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_flag+=("--local-port=")
    must_have_one_flag+=("--remote=")
    must_have_one_flag+=("-r")
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_proxy()
{
    last_command="gocat_proxy"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--backends=")
    two_word_flags+=("--backends")
    local_nonpersistent_flags+=("--backends")
    local_nonpersistent_flags+=("--backends=")
    flags+=("--cert=")
    two_word_flags+=("--cert")
    local_nonpersistent_flags+=("--cert")
    local_nonpersistent_flags+=("--cert=")
    flags+=("--health-check=")
    two_word_flags+=("--health-check")
    local_nonpersistent_flags+=("--health-check")
    local_nonpersistent_flags+=("--health-check=")
    flags+=("--insecure-backend")
    local_nonpersistent_flags+=("--insecure-backend")
    flags+=("--key=")
    two_word_flags+=("--key")
    local_nonpersistent_flags+=("--key")
    local_nonpersistent_flags+=("--key=")
    flags+=("--lb-algorithm=")
    two_word_flags+=("--lb-algorithm")
    local_nonpersistent_flags+=("--lb-algorithm")
    local_nonpersistent_flags+=("--lb-algorithm=")
    flags+=("--listen=")
    two_word_flags+=("--listen")
    two_word_flags+=("-l")
    local_nonpersistent_flags+=("--listen")
    local_nonpersistent_flags+=("--listen=")
    local_nonpersistent_flags+=("-l")
    flags+=("--log-requests")
    local_nonpersistent_flags+=("--log-requests")
    flags+=("--max-connections=")
    two_word_flags+=("--max-connections")
    local_nonpersistent_flags+=("--max-connections")
    local_nonpersistent_flags+=("--max-connections=")
    flags+=("--modify-headers")
    local_nonpersistent_flags+=("--modify-headers")
    flags+=("--ssl")
    local_nonpersistent_flags+=("--ssl")
    flags+=("--target=")
    two_word_flags+=("--target")
    local_nonpersistent_flags+=("--target")
    local_nonpersistent_flags+=("--target=")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_scan()
{
    last_command="gocat_scan"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--concurrency=")
    two_word_flags+=("--concurrency")
    local_nonpersistent_flags+=("--concurrency")
    local_nonpersistent_flags+=("--concurrency=")
    flags+=("--open")
    local_nonpersistent_flags+=("--open")
    flags+=("--ports=")
    two_word_flags+=("--ports")
    local_nonpersistent_flags+=("--ports")
    local_nonpersistent_flags+=("--ports=")
    flags+=("--scan-verbose")
    local_nonpersistent_flags+=("--scan-verbose")
    flags+=("--verbose-scan")
    local_nonpersistent_flags+=("--verbose-scan")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_script_info()
{
    last_command="gocat_script_info"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_script_list()
{
    last_command="gocat_script_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--detailed")
    local_nonpersistent_flags+=("--detailed")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_script_run()
{
    last_command="gocat_script_run"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--args=")
    two_word_flags+=("--args")
    two_word_flags+=("-a")
    local_nonpersistent_flags+=("--args")
    local_nonpersistent_flags+=("--args=")
    local_nonpersistent_flags+=("-a")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--verbose")
    flags+=("-v")
    local_nonpersistent_flags+=("--verbose")
    local_nonpersistent_flags+=("-v")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_script_validate()
{
    last_command="gocat_script_validate"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_script()
{
    last_command="gocat_script"

    command_aliases=()

    commands=()
    commands+=("info")
    commands+=("list")
    commands+=("run")
    commands+=("validate")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_serve()
{
    last_command="gocat_serve"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--auth=")
    two_word_flags+=("--auth")
    local_nonpersistent_flags+=("--auth")
    local_nonpersistent_flags+=("--auth=")
    flags+=("--dir=")
    two_word_flags+=("--dir")
    local_nonpersistent_flags+=("--dir")
    local_nonpersistent_flags+=("--dir=")
    flags+=("--host=")
    two_word_flags+=("--host")
    local_nonpersistent_flags+=("--host")
    local_nonpersistent_flags+=("--host=")
    flags+=("--port=")
    two_word_flags+=("--port")
    local_nonpersistent_flags+=("--port")
    local_nonpersistent_flags+=("--port=")
    flags+=("--prefix=")
    two_word_flags+=("--prefix")
    local_nonpersistent_flags+=("--prefix")
    local_nonpersistent_flags+=("--prefix=")
    flags+=("--quiet")
    local_nonpersistent_flags+=("--quiet")
    flags+=("--upload")
    local_nonpersistent_flags+=("--upload")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_session_agent()
{
    last_command="gocat_session_agent"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--host=")
    two_word_flags+=("--host")
    local_nonpersistent_flags+=("--host")
    local_nonpersistent_flags+=("--host=")
    flags+=("--port=")
    two_word_flags+=("--port")
    local_nonpersistent_flags+=("--port")
    local_nonpersistent_flags+=("--port=")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_session_cleanup()
{
    last_command="gocat_session_cleanup"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_session_detach()
{
    last_command="gocat_session_detach"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_session_download()
{
    last_command="gocat_session_download"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_session_exec()
{
    last_command="gocat_session_exec"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_session_info()
{
    last_command="gocat_session_info"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_session_interact()
{
    last_command="gocat_session_interact"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_session_kill()
{
    last_command="gocat_session_kill"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_session_killall()
{
    last_command="gocat_session_killall"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_session_list()
{
    last_command="gocat_session_list"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_session_modules()
{
    last_command="gocat_session_modules"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_session_portfwd()
{
    last_command="gocat_session_portfwd"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--agent-host=")
    two_word_flags+=("--agent-host")
    local_nonpersistent_flags+=("--agent-host")
    local_nonpersistent_flags+=("--agent-host=")
    flags+=("--agent-port=")
    two_word_flags+=("--agent-port")
    local_nonpersistent_flags+=("--agent-port")
    local_nonpersistent_flags+=("--agent-port=")
    flags+=("--bind=")
    two_word_flags+=("--bind")
    local_nonpersistent_flags+=("--bind")
    local_nonpersistent_flags+=("--bind=")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_session_run()
{
    last_command="gocat_session_run"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_session_upgrade()
{
    last_command="gocat_session_upgrade"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_session_upload()
{
    last_command="gocat_session_upload"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_session()
{
    last_command="gocat_session"

    command_aliases=()

    commands=()
    commands+=("agent")
    commands+=("cleanup")
    commands+=("detach")
    commands+=("download")
    commands+=("exec")
    commands+=("info")
    commands+=("interact")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("attach")
        aliashash["attach"]="interact"
        command_aliases+=("use")
        aliashash["use"]="interact"
    fi
    commands+=("kill")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("remove")
        aliashash["remove"]="kill"
        command_aliases+=("rm")
        aliashash["rm"]="kill"
    fi
    commands+=("killall")
    commands+=("list")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("ls")
        aliashash["ls"]="list"
    fi
    commands+=("modules")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("mods")
        aliashash["mods"]="modules"
        command_aliases+=("module")
        aliashash["module"]="modules"
    fi
    commands+=("portfwd")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("forward")
        aliashash["forward"]="portfwd"
        command_aliases+=("pf")
        aliashash["pf"]="portfwd"
    fi
    commands+=("run")
    commands+=("upgrade")
    commands+=("upload")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_sniffer()
{
    last_command="gocat_sniffer"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_stabilize()
{
    last_command="gocat_stabilize"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--method=")
    two_word_flags+=("--method")
    local_nonpersistent_flags+=("--method")
    local_nonpersistent_flags+=("--method=")
    flags+=("--resize")
    local_nonpersistent_flags+=("--resize")
    flags+=("--shell=")
    two_word_flags+=("--shell")
    local_nonpersistent_flags+=("--shell")
    local_nonpersistent_flags+=("--shell=")
    flags+=("--upgrade")
    local_nonpersistent_flags+=("--upgrade")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_transfer()
{
    last_command="gocat_transfer"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--buffer=")
    two_word_flags+=("--buffer")
    local_nonpersistent_flags+=("--buffer")
    local_nonpersistent_flags+=("--buffer=")
    flags+=("--checksum")
    local_nonpersistent_flags+=("--checksum")
    flags+=("--compress")
    local_nonpersistent_flags+=("--compress")
    flags+=("--file=")
    two_word_flags+=("--file")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--file")
    local_nonpersistent_flags+=("--file=")
    local_nonpersistent_flags+=("-f")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output")
    local_nonpersistent_flags+=("--output=")
    local_nonpersistent_flags+=("-o")
    flags+=("--progress")
    local_nonpersistent_flags+=("--progress")
    flags+=("--resume")
    local_nonpersistent_flags+=("--resume")
    flags+=("--transfer-timeout=")
    two_word_flags+=("--transfer-timeout")
    local_nonpersistent_flags+=("--transfer-timeout")
    local_nonpersistent_flags+=("--transfer-timeout=")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_tunnel()
{
    last_command="gocat_tunnel"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--compression")
    local_nonpersistent_flags+=("--compression")
    flags+=("--dynamic")
    local_nonpersistent_flags+=("--dynamic")
    flags+=("--insecure-hostkey")
    local_nonpersistent_flags+=("--insecure-hostkey")
    flags+=("--key=")
    two_word_flags+=("--key")
    local_nonpersistent_flags+=("--key")
    local_nonpersistent_flags+=("--key=")
    flags+=("--local=")
    two_word_flags+=("--local")
    local_nonpersistent_flags+=("--local")
    local_nonpersistent_flags+=("--local=")
    flags+=("--password=")
    two_word_flags+=("--password")
    local_nonpersistent_flags+=("--password")
    local_nonpersistent_flags+=("--password=")
    flags+=("--remote=")
    two_word_flags+=("--remote")
    local_nonpersistent_flags+=("--remote")
    local_nonpersistent_flags+=("--remote=")
    flags+=("--reverse")
    local_nonpersistent_flags+=("--reverse")
    flags+=("--ssh=")
    two_word_flags+=("--ssh")
    local_nonpersistent_flags+=("--ssh")
    local_nonpersistent_flags+=("--ssh=")
    flags+=("--timeout=")
    two_word_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout")
    local_nonpersistent_flags+=("--timeout=")
    flags+=("--user=")
    two_word_flags+=("--user")
    local_nonpersistent_flags+=("--user")
    local_nonpersistent_flags+=("--user=")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_flag+=("--ssh=")
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_unix_connect()
{
    last_command="gocat_unix_connect"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--buffer=")
    two_word_flags+=("--buffer")
    local_nonpersistent_flags+=("--buffer")
    local_nonpersistent_flags+=("--buffer=")
    flags+=("--type=")
    two_word_flags+=("--type")
    local_nonpersistent_flags+=("--type")
    local_nonpersistent_flags+=("--type=")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_unix_echo()
{
    last_command="gocat_unix_echo"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--permissions=")
    two_word_flags+=("--permissions")
    local_nonpersistent_flags+=("--permissions")
    local_nonpersistent_flags+=("--permissions=")
    flags+=("--remove")
    local_nonpersistent_flags+=("--remove")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_unix_listen()
{
    last_command="gocat_unix_listen"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--buffer=")
    two_word_flags+=("--buffer")
    local_nonpersistent_flags+=("--buffer")
    local_nonpersistent_flags+=("--buffer=")
    flags+=("--permissions=")
    two_word_flags+=("--permissions")
    local_nonpersistent_flags+=("--permissions")
    local_nonpersistent_flags+=("--permissions=")
    flags+=("--remove")
    local_nonpersistent_flags+=("--remove")
    flags+=("--type=")
    two_word_flags+=("--type")
    local_nonpersistent_flags+=("--type")
    local_nonpersistent_flags+=("--type=")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_unix()
{
    last_command="gocat_unix"

    command_aliases=()

    commands=()
    commands+=("connect")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("c")
        aliashash["c"]="connect"
    fi
    commands+=("echo")
    commands+=("listen")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_verify()
{
    last_command="gocat_verify"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_version()
{
    last_command="gocat_version"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_websocket_connect()
{
    last_command="gocat_websocket_connect"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--compress")
    local_nonpersistent_flags+=("--compress")
    flags+=("--origin=")
    two_word_flags+=("--origin")
    local_nonpersistent_flags+=("--origin")
    local_nonpersistent_flags+=("--origin=")
    flags+=("--ping-interval=")
    two_word_flags+=("--ping-interval")
    local_nonpersistent_flags+=("--ping-interval")
    local_nonpersistent_flags+=("--ping-interval=")
    flags+=("--read-buffer=")
    two_word_flags+=("--read-buffer")
    local_nonpersistent_flags+=("--read-buffer")
    local_nonpersistent_flags+=("--read-buffer=")
    flags+=("--write-buffer=")
    two_word_flags+=("--write-buffer")
    local_nonpersistent_flags+=("--write-buffer")
    local_nonpersistent_flags+=("--write-buffer=")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_websocket_echo()
{
    last_command="gocat_websocket_echo"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--allow-all-origins")
    local_nonpersistent_flags+=("--allow-all-origins")
    flags+=("--allowed-origins=")
    two_word_flags+=("--allowed-origins")
    local_nonpersistent_flags+=("--allowed-origins")
    local_nonpersistent_flags+=("--allowed-origins=")
    flags+=("--compress")
    local_nonpersistent_flags+=("--compress")
    flags+=("--path=")
    two_word_flags+=("--path")
    local_nonpersistent_flags+=("--path")
    local_nonpersistent_flags+=("--path=")
    flags+=("--port=")
    two_word_flags+=("--port")
    local_nonpersistent_flags+=("--port")
    local_nonpersistent_flags+=("--port=")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_websocket_server()
{
    last_command="gocat_websocket_server"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--allow-all-origins")
    local_nonpersistent_flags+=("--allow-all-origins")
    flags+=("--allowed-origins=")
    two_word_flags+=("--allowed-origins")
    local_nonpersistent_flags+=("--allowed-origins")
    local_nonpersistent_flags+=("--allowed-origins=")
    flags+=("--compress")
    local_nonpersistent_flags+=("--compress")
    flags+=("--max-message-size=")
    two_word_flags+=("--max-message-size")
    local_nonpersistent_flags+=("--max-message-size")
    local_nonpersistent_flags+=("--max-message-size=")
    flags+=("--path=")
    two_word_flags+=("--path")
    local_nonpersistent_flags+=("--path")
    local_nonpersistent_flags+=("--path=")
    flags+=("--ping-interval=")
    two_word_flags+=("--ping-interval")
    local_nonpersistent_flags+=("--ping-interval")
    local_nonpersistent_flags+=("--ping-interval=")
    flags+=("--pong-timeout=")
    two_word_flags+=("--pong-timeout")
    local_nonpersistent_flags+=("--pong-timeout")
    local_nonpersistent_flags+=("--pong-timeout=")
    flags+=("--port=")
    two_word_flags+=("--port")
    local_nonpersistent_flags+=("--port")
    local_nonpersistent_flags+=("--port=")
    flags+=("--read-buffer=")
    two_word_flags+=("--read-buffer")
    local_nonpersistent_flags+=("--read-buffer")
    local_nonpersistent_flags+=("--read-buffer=")
    flags+=("--write-buffer=")
    two_word_flags+=("--write-buffer")
    local_nonpersistent_flags+=("--write-buffer")
    local_nonpersistent_flags+=("--write-buffer=")
    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_websocket()
{
    last_command="gocat_websocket"

    command_aliases=()

    commands=()
    commands+=("connect")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("c")
        aliashash["c"]="connect"
        command_aliases+=("client")
        aliashash["client"]="connect"
    fi
    commands+=("echo")
    commands+=("server")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_gocat_root_command()
{
    last_command="gocat"

    command_aliases=()

    commands=()
    commands+=("benchmark")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("bench")
        aliashash["bench"]="benchmark"
        command_aliases+=("stress")
        aliashash["stress"]="benchmark"
    fi
    commands+=("broker")
    commands+=("chat")
    commands+=("completion")
    commands+=("connect")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("c")
        aliashash["c"]="connect"
    fi
    commands+=("console")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("menu")
        aliashash["menu"]="console"
        command_aliases+=("repl")
        aliashash["repl"]="console"
    fi
    commands+=("convert")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("conv")
        aliashash["conv"]="convert"
        command_aliases+=("protocol-convert")
        aliashash["protocol-convert"]="convert"
    fi
    commands+=("distributed")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("cluster")
        aliashash["cluster"]="distributed"
        command_aliases+=("dist")
        aliashash["dist"]="distributed"
    fi
    commands+=("dns-tunnel")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("dns")
        aliashash["dns"]="dns-tunnel"
        command_aliases+=("dnstun")
        aliashash["dnstun"]="dns-tunnel"
    fi
    commands+=("doctor")
    commands+=("help")
    commands+=("interfaces")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("ifaces")
        aliashash["ifaces"]="interfaces"
        command_aliases+=("ifconfig")
        aliashash["ifconfig"]="interfaces"
        command_aliases+=("ifs")
        aliashash["ifs"]="interfaces"
    fi
    commands+=("listen")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("l")
        aliashash["l"]="listen"
    fi
    commands+=("mcp")
    commands+=("metrics")
    commands+=("multi-listen")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("ml")
        aliashash["ml"]="multi-listen"
        command_aliases+=("multi")
        aliashash["multi"]="multi-listen"
    fi
    commands+=("payload")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("gen")
        aliashash["gen"]="payload"
        command_aliases+=("generate")
        aliashash["generate"]="payload"
        command_aliases+=("payloads")
        aliashash["payloads"]="payload"
    fi
    commands+=("portforward")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("forward")
        aliashash["forward"]="portforward"
        command_aliases+=("pf")
        aliashash["pf"]="portforward"
    fi
    commands+=("proxy")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("p")
        aliashash["p"]="proxy"
        command_aliases+=("reverse-proxy")
        aliashash["reverse-proxy"]="proxy"
    fi
    commands+=("scan")
    commands+=("script")
    commands+=("serve")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("fileserver")
        aliashash["fileserver"]="serve"
        command_aliases+=("fs")
        aliashash["fs"]="serve"
        command_aliases+=("http")
        aliashash["http"]="serve"
    fi
    commands+=("session")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("sess")
        aliashash["sess"]="session"
        command_aliases+=("sessions")
        aliashash["sessions"]="session"
    fi
    commands+=("sniffer")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("capture")
        aliashash["capture"]="sniffer"
        command_aliases+=("sniff")
        aliashash["sniff"]="sniffer"
    fi
    commands+=("stabilize")
    commands+=("transfer")
    commands+=("tunnel")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("ssh-tunnel")
        aliashash["ssh-tunnel"]="tunnel"
        command_aliases+=("tun")
        aliashash["tun"]="tunnel"
    fi
    commands+=("unix")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("uds")
        aliashash["uds"]="unix"
    fi
    commands+=("verify")
    commands+=("version")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("v")
        aliashash["v"]="version"
    fi
    commands+=("websocket")
    if [[ -z "${BASH_VERSION:-}" || "${BASH_VERSINFO[0]:-}" -gt 3 ]]; then
        command_aliases+=("ws")
        aliashash["ws"]="websocket"
    fi

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--debug")
    flags+=("--keep-open")
    flags+=("-k")
    flags+=("--listen")
    flags+=("-l")
    flags+=("--port-range=")
    two_word_flags+=("--port-range")
    flags+=("--profile=")
    two_word_flags+=("--profile")
    flags+=("--scan")
    flags+=("-z")
    flags+=("--ssl")
    flags+=("--udp")
    flags+=("-u")
    flags+=("--verbose")
    flags+=("-v")
    flags+=("--wait=")
    two_word_flags+=("--wait")
    two_word_flags+=("-w")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

__start_gocat()
{
    local cur prev words cword split
    declare -A flaghash 2>/dev/null || :
    declare -A aliashash 2>/dev/null || :
    if declare -F _init_completion >/dev/null 2>&1; then
        _init_completion -s || return
    else
        __gocat_init_completion -n "=" || return
    fi

    local c=0
    local flag_parsing_disabled=
    local flags=()
    local two_word_flags=()
    local local_nonpersistent_flags=()
    local flags_with_completion=()
    local flags_completion=()
    local commands=("gocat")
    local command_aliases=()
    local must_have_one_flag=()
    local must_have_one_noun=()
    local has_completion_function=""
    local last_command=""
    local nouns=()
    local noun_aliases=()

    __gocat_handle_word
}

if [[ $(type -t compopt) = "builtin" ]]; then
    complete -o default -F __start_gocat gocat
else
    complete -o default -o nospace -F __start_gocat gocat
fi

# ex: ts=4 sw=4 et filetype=sh
