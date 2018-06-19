# If not found, return nothing.
if response.status_code == 404:
    return None

# If no content, return nothing.
if response.status_code == 204:
    return None

if response.status_code == 403:
    return None

try:
    result = response.json()
except getattr(json.decoder, 'JSONDecodeError', ValueError) as inst:
    module.fail_json(msg="Invalid JSON response with error: %s" % inst)

if navigate_hash(result, ['error', 'code']):
    module.fail_json(msg=navigate_hash(result, ['error', 'details']))

return result
