#!/bin/sh

BASE_URL="http://localhost:8090"

CMD=$1
shift

case "$CMD" in
hello)
	curl -s -X GET "$BASE_URL/hello"
	;;

login)
	USERNAME=$1
	PASSWORD=$2
	curl -s -X POST "$BASE_URL/login" \
		-H "Content-Type: application/json" \
		-d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}"
	;;

validate)
	USERNAME=$1
	OTP=$2
	curl -s -X POST "$BASE_URL/validate" \
		-H "Content-Type: application/json" \
		-d "{\"username\":\"$USERNAME\",\"otp\":\"$OTP\"}"
	;;

protected)
	TOKEN=$1
	curl -s -X GET "$BASE_URL/protected" \
		-H "Authorization: Bearer $TOKEN"
	;;

*)
	echo "Usage:"
	echo "  $0 hello"
	echo "  $0 login <username> <password>"
	echo "  $0 validate <username> <otp>"
	echo "  $0 protected <token>"
	exit 1
	;;
esac
