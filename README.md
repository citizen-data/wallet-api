# wallet-api

```
POST    /wallet                                 Create wallet
GET     /wallet/{walletID}                      Get wallet metadata
POST    /wallet/{walletID}/data                 Add data item
GET     /wallet/{walletID}/data                 Get list of data (summary data)
GET     /wallet/{walletID}/{refID}              Get history for dataItem (including encrypted data)
GET     /wallet/{walletID}/{refID}/latest       Get latest version of dataItem (including encrypted data)
GET     /wallet/{walletID}/{refID}/{dataHash}   Get specific version of dataItem  (including encrypted data)
```


## Required Headers
```
x-api-key:  		tenant api key (would be embedded in app)
x-api-timestamp:	time of request 2006-01-02T15:04:05.000Z (must be within 10s)
x-api-signature: 	base64(PKCS1v15(sha256("urlpath|body|x-api-timestamp")))
```


## Build
```$xslt
make build
```

## Test
```$xslt
make test
```

## Deploy
```$xslt
make deploy
```