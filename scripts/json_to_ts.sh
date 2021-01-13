#!/bin/bash
set -e

json=$1
model_name=$2

if [ -z "$json" ];
then
    echo "No JSON specified"
    exit 1
fi
if [ -z "$model_name" ];
then
    echo "No model name specified"
    exit 1
fi

mkdir -p $GOPATH/src/tmp_models
echo "Using $GOPATH/src/tmp_models ad temporary models package"
cat $model_name.json | gojson -pkg tmp_models -name=$model_name > $GOPATH/src/tmp_models/$model_name.go
tscriptify -package=tmp_models -target $model_name.ts $model_name
echo "Saved to $model_name"
