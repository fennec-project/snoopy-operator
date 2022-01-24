


URL=https://github.com/operator-framework/operator-sdk/releases/download/v1.16.0/operator-sdk_darwin_amd64

cd /usr/local/bin && rm operator-sdk && curl -LO $URL && mv operator-sdk_darwin_amd64 operator-sdk && chmod +x operator-sdk;

cd -;