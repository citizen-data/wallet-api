# wallet-api

```
POST    /wallet                                     Create wallet
POST    /wallet/{walletID}                          Add data item
GET     /wallet/{walletID}                          Get list of data (summary data)
GET     /wallet/{walletID}/data/{refID}             Get history for dataItem (w/encrypted data)
GET     /wallet/{walletID}/data/{refID}/latest      Get latest version of dataItem (w/encrypted data)
GET     /wallet/{walletID}/data/{refID}/{dataHash}  Get specific version of dataItem (w/encrypted data)
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