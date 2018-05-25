#!/usr/bin/env bash

C_RED="\033[31;01m"
C_GREEN="\033[32;01m"
C_YELLOW="\033[33;01m"
C_BLUE="\033[34;01m"
C_PINK="\033[35;01m"
C_CYAN="\033[36;01m"
C_NO="\033[0m"

################################################################################
### FUNCTIONS ##################################################################
################################################################################

function	usage()
{
	case ${1} in
		balance)
			printf "usage: balance\n"
			;;
		spend)
			printf "usage: spend AMOUNT LABEL NEW_OWNER\n"
			;;
		*)
			;;
	esac
}

function	ope_balance()
{
	if [[ ${#} -gt 0 ]]; then
		printf "balance does not require any argument.\n"
		usage balance; return
	fi
	printf "balance %s\n" ${1}
}

function	ope_spend()
{
	if [[ ${#} -lt 3 ]]; then
		printf "spend requires tree arguments.\n"
		usage spend ; return
	elif [[ ${#} -gt 3 ]]; then
		printf "spend requires only tree arguments.\n"
		usage spend ; return
	fi
	printf "spend %s %s %s\n" "${1}" "${2}" "${3}"
}

function	get_input()
{
	### READ INPUT
	#printf "${C_CYAN}%s${C_NO}\$ " "wsh"
	read -e -p "wsh$ " -a argv
	argc=${#argv[@]}
	
	case ${argv[0]} in
		balance|spend)	ope_${argv[0]} ${argv[@]:1} ;;
		exit|quit)		exit 0 ;;
		"")				;;
		*)				usage ;;
	esac
}

################################################################################
### MAIN #######################################################################
################################################################################

while [[ 1 ]]; do
	get_input
done
