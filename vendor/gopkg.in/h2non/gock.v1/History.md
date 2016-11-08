## 1.0.1 / 19-04-2016

- fix(#9): missing URL query param matcher.

## 1.0.0 / 19-04-2016

- feat(version): first major version release.

## 0.1.6 / 19-04-2016

- fix(#7): if error configured, RoundTripper should reply with `nil` response.

## 0.1.5 / 09-04-2016

- feat(#5): support `ReplyFunc` for convenience.

## 0.1.4 / 16-03-2016

- feat(api): add `IsDone()` method.
- fix(responder): return mock error if present.
- feat(#4): support define request/response body from file disk.

## 0.1.3 / 09-03-2016

- feat(matcher): add content type matcher helper method supporting aliases. 
- feat(interceptor): add function to restore HTTP client transport.
- feat(matcher): add URL scheme matcher function.
- fix(request): ignore base slash path.
- feat(api): add Off() method for easier restore and clean up.
- feat(store): add public API for pending mocks.

## 0.1.2 / 04-03-2016

- fix(matcher): body matchers no used by default.
- feat(matcher): add matcher factories for multiple cases. 

## 0.1.1 / 04-03-2016

- fix(params): persist query params accordingly. 

## 0.1.0 / 02-03-2016

- First release.
