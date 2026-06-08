#!/bin/bash
# Acceptance test script for FamilyOps AI Knowledge Base
# Runs 45 questions across 5 domains and saves results

API_BASE="${API_BASE:-http://127.0.0.1:18081}"
TOKEN="${TOKEN:-029ce14b847a4dd28cdfcb5782d214eb50360a61a0ff4282936ef89580781ad1}"
OUTPUT_DIR="${OUTPUT_DIR:-./acceptance_results}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

mkdir -p "$OUTPUT_DIR"
RESULT_FILE="$OUTPUT_DIR/results_${TIMESTAMP}.txt"
SUMMARY_FILE="$OUTPUT_DIR/summary_${TIMESTAMP}.txt"

# Create a conversation for the test
CID=$(curl -s -X POST "$API_BASE/api/chat/conversations" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"title\":\"Acceptance Test $TIMESTAMP\"}" | python3 -c "import sys,json; print(json.load(sys.stdin).get('id',''))")

if [ -z "$CID" ]; then
  echo "FATAL: Failed to create conversation"
  exit 1
fi

echo "Test conversation: $CID"
echo "Results: $RESULT_FILE"

# Domain questions
declare -A DOMAINS
DOMAINS=(
  ["finance"]="财务领域"
  ["health"]="健康领域"
  ["work"]="工作领域"
  ["family"]="家庭事务领域"
  ["mixed"]="混合领域"
)

# All 45 questions
QUESTIONS=(
  # Finance (10)
  "finance|家里现在一共有多少钱？"
  "finance|这个月花了多少？跟上个月比呢？"
  "finance|房贷还剩多少？每个月还多少？"
  "finance|家庭净资产大概是多少？"
  "finance|这个月最大的开支是什么？"
  "finance|爷爷最近有什么医疗支出吗？"
  "finance|我们每月固定支出大概是多少？"
  "finance|基金投了多少钱？"
  "finance|今年上半年收入大概多少？"
  "finance|有什么可以省钱的地方吗？"
  # Health (10)
  "health|妈最近血压怎么样？"
  "health|爸在吃什么药？"
  "health|一鸣的身高体重是多少？"
  "health|爷爷的血糖控制得怎么样？"
  "health|家里谁最近在运动？"
  "health|一鸣的视力怎么样了？"
  "health|奶奶的膝盖好点了吗？"
  "health|妈的血压有没有在好转？"
  "health|一鸣有什么要注意的健康问题吗？"
  "health|全家人的体重都是多少？"
  # Work (10)
  "work|最近有什么重要的deadline？"
  "work|志豪的工作项目进展怎么样？"
  "work|小婉这个月有什么工作安排？"
  "work|最紧急的工作事项是什么？"
  "work|志豪每周有什么固定会议？"
  "work|小婉的期末出卷准备得怎么样了？"
  "work|有什么已经完成的工作里程碑吗？"
  "work|志豪的博客更新了吗？"
  "work|两人的通勤情况是怎样的？"
  "work|下个月工作方面有什么重点？"
  # Family (10)
  "family|这周有什么安排？"
  "family|下周三家长会谁去？"
  "family|暑假孩子的看护安排好了吗？"
  "family|外婆生日什么时候？准备了吗？"
  "family|最近有什么家务要做？"
  "family|一鸣什么时候去复查视力？"
  "family|车子该保养了吗？"
  "family|父亲节礼物准备了吗？"
  "family|孩子们在哪个学校上学？"
  "family|暑假全家出游计划得怎么样了？"
  # Mixed (5)
  "mixed|最近家里整体情况怎么样？"
  "mixed|这个月家庭有什么大事需要注意？"
  "mixed|志豪最近忙不忙？工作和健康怎么样？"
  "mixed|一鸣最近各方面情况怎么样？"
  "mixed|帮我看一下下个月需要提前准备什么"
)

PASS=0
FAIL=0
TOTAL=45
CURRENT=0

echo "=== FamilyOps Acceptance Test ($TIMESTAMP) ===" > "$RESULT_FILE"
echo "" >> "$RESULT_FILE"

for Q in "${QUESTIONS[@]}"; do
  DOMAIN="${Q%%|*}"
  QUESTION="${Q#*|}"
  CURRENT=$((CURRENT + 1))

  echo -n "[$CURRENT/$TOTAL] ${DOMAIN}: $QUESTION ... "

  # Send message
  START_TIME=$(date +%s)
  RESPONSE=$(curl -s -X POST "$API_BASE/api/chat/conversations/$CID/messages" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"content\":\"$QUESTION\"}" 2>&1)
  END_TIME=$(date +%s)
  ELAPSED=$((END_TIME - START_TIME))

  ANSWER=$(echo "$RESPONSE" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('content','ERROR_NO_CONTENT'))" 2>/dev/null)

  if [ -z "$ANSWER" ] || [[ "$ANSWER" == ERROR_* ]] || [[ "$ANSWER" == *"AI回答失败"* ]]; then
    echo "FAIL (${ELAPSED}s)"
    FAIL=$((FAIL + 1))
    STATUS="FAIL"
  else
    echo "OK (${ELAPSED}s)"
    PASS=$((PASS + 1))
    STATUS="PASS"
  fi

  # Write result
  {
    echo "--- [$CURRENT/$TOTAL] $DOMAIN | $STATUS | ${ELAPSED}s ---"
    echo "Q: $QUESTION"
    echo "A: $ANSWER"
    echo ""
  } >> "$RESULT_FILE"
done

{
  echo "=== Summary ==="
  echo "Total: $TOTAL"
  echo "Pass: $PASS"
  echo "Fail: $FAIL"
  echo "Pass Rate: $(( PASS * 100 / TOTAL ))%"
  echo "Results: $RESULT_FILE"
} | tee "$SUMMARY_FILE"

echo ""
echo "Done! Results saved to $RESULT_FILE"
echo "Summary saved to $SUMMARY_FILE"
