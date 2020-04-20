buildoperator:
	./build-operator.sh
buildqueryjob:
	./build-queryjob.sh

buildall: buildqueryjob buildoperator