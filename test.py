import requests

# The URL you want to make a request to
url = "http://httpbin.org/status/200"#"https://openai.com/"#"http://httpbin.org/status/200"

# Proxy configuration: Adjust the port if your proxy server uses a different one
proxies = {
    "http": "http://localhost:8080",
    "https": "http://localhost:8080",
}

# Make the request through the proxy
response = requests.get(url, proxies=proxies)

print("with proxy")
# Print the response status code
print(f"Response status code: {response.status_code}")
original=response.content

print("vanila")
print(f"Response status code: {response.status_code}")
assert original==response.content
