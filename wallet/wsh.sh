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
			printf "usage: spend AMOUNT OWNER LABEL\n"
			;;
		*)
			printf "usage mode [ARGUMENTS]\n"
			printf "  - balance: balance\n"
			printf "  - spend:   spend AMOUNT OWNER LABEL\n"
			;;
	esac
}

function	ope_balance()
{
	if [[ ${#} -gt 0 ]]; then
		printf "balance does not require any argument.\n"
		usage balance; return
	fi
	./peer balance
}

function	ope_spend()
{
	amount=$1
	owner=$2
	label=$3

	if [[ ${#} -gt 3 ]]; then
		printf "spend requires only tree arguments.\n"
		usage spend ; return
	fi
	while [[ -z ${amount} ]]; do read -e -p "spend-amount$ " amount; done
	while [[ -z ${owner} ]]; do read -e -p "spend-owner$ " owner; done
	while [[ -z ${label} ]]; do read -e -p "spend-label$ " label; done
	./peer spend ${amount} ${owner} ${label}
}

function	get_input()
{
	read -e -p "wsh:$(basename $(pwd))$ " -a argv
	
	case ${argv[0]} in
		balance|spend)	ope_${argv[0]} ${argv[@]:1} ;;
		exit|quit)		exit 0 ;;
		*)				${argv[@]} ;;
	esac
}

################################################################################
### MAIN #######################################################################
################################################################################

while [[ 1 ]]; do
	get_input
done
