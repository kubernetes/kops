import jwt
import datetime
import hashlib
import httplib
import time

method = 'GET'
uri = '/clusters'
secret = 'My Secret'

claims = {}

claims['iss'] = 'admin'

# Issued at time
claims['iat'] = datetime.datetime.utcnow()

# Expiration time
claims['exp'] = datetime.datetime.utcnow() \
	+ datetime.timedelta(seconds=1)

# URI tampering protection
claims['qsh'] = hashlib.sha256(method + '&' + uri).hexdigest()

token = jwt.encode(claims, secret, algorithm='HS256')
h = httplib.HTTPConnection('localhost:8080')
headers = {}
headers['Authorization'] = 'bearer ' + token

h.request(method, uri, headers=headers)
r = h.getresponse()
print r.read()
