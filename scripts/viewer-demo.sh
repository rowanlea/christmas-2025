#!/bin/sh
# Posts a demo publication to a running viewer daemon so you can watch it in a browser.
# Usage: HTH_VIEWER_PORT=47999 go run ./cmd/hth viewer   (in another terminal), open http://127.0.0.1:$PORT/
# then: HTH_VIEWER_PORT=47999 scripts/viewer-demo.sh
set -e
PORT="${HTH_VIEWER_PORT:-47474}"
BASE="http://127.0.0.1:$PORT"

post() {
  curl -fsS -X POST -H "X-HTH: 1" -H "Content-Type: application/json" -d "$2" "$BASE$1"
}

# Size 4 = 10 bulbs, size 16 = 136 bulbs.
EXP4=RGBWKRGBWK
ANS4=RGBWKGGBWB
ANS16=$(printf 'R%.0s' $(seq 1 136))

post /publish '{"id":"demo","kind":"test","cards":[
 {"id":"a","title":"example 1 (4 rows)","size":4,"answer":"'$ANS4'","expected":"'$EXP4'","summary":"2 of 10 bulbs differ","frameCount":4,"intervalMs":100},
 {"id":"b","title":"example 2 (16 rows)","size":16,"answer":"'$ANS16'","expected":"'$ANS16'","summary":"all bulbs match","frameCount":3,"intervalMs":100}]}'

# Card a: frames out of order, index 1 skipped for a gap, index 3 preceded by a skip caption.
post /frame '{"publication":"demo","card":"a","index":2,"line":"RRRRKKKKWW"}'
post /frame '{"publication":"demo","card":"a","index":0,"line":"RKKKKKKKKK"}'
sleep 1
post /frame '{"publication":"demo","card":"a","index":3,"line":"RGBWKGGBWB","skippedBefore":999999900}'
post /frame '{"publication":"demo","card":"b","index":0,"line":"'$ANS16'"}'
sleep 1
post /frame '{"publication":"demo","card":"b","index":1,"line":"'$ANS16'"}'
post /done '{"publication":"demo","card":"a","skippedAtEnd":123456}'
post /done '{"publication":"demo","card":"b","skippedAtEnd":0}'
echo "demo posted to $BASE"
