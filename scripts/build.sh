DIR="./app/builds"
if [ ! -d "$DIR" ]; then
   echo "Creating directory: $DIR"
    mkdir ./app/builds
fi

go build -o ./app/builds/read ./app/read
go build -o ./app/builds/audit ./app/audit
go build -o ./app/builds/transfer ./app/transfer
go build -o ./app/builds/transactions ./app/transactions