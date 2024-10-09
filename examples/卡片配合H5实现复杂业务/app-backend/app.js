const express = require("express");
const cors = require("cors");
const axios = require("axios");
const { program } = require("commander");
const { DWClient } = require("dingtalk-stream");
const app = express();

app.use(cors());
app.use(express.json()); // 解析 JSON 数据

app.post("/api/submit", (req, res) => {
  // req.body demo: { stars: 3, tags: ["评价 A", "评价 C"], suggestion: "一段建议", cardInstanceId: "123456789" }
  console.log(req.body); // 输出前端发送的数据
  // 更新卡片
  const { cardInstanceId, ...evaluate } = req.body;
  const cardData = { evaluate, evaluate_done: true };
  updateCard({ cardInstanceId, cardData });
  res.json({ message: "success" });
});

app.listen(5000, () => {
  console.log("Server is running on http://localhost:5000");
});

// 更新卡片
program
  .requiredOption(
    "--clientId <Client ID>",
    "your client id, AppKey or SuiteKey",
    process.env.DINGTALK_APP_CLIENT_ID
  )
  .requiredOption(
    "--clientSecret <Client Secret>",
    "your client secret, AppSecret or SuiteSecret",
    process.env.DINGTALK_APP_CLIENT_SECRET
  )
  .parse();
const options = program.opts();

const client = new DWClient({
  clientId: options.clientId,
  clientSecret: options.clientSecret,
  debug: false, // 调试模式，开启后可以看到更多详细日志
});

const convertJSONValuesToString = (obj) => {
  const newObj = {};
  for (const key in obj) {
    const value = obj[key];
    if (obj.hasOwnProperty(key) && value != null) {
      if (typeof value === "string") {
        newObj[key] = value;
      } else {
        newObj[key] = JSON.stringify(value);
      }
    }
  }
  return newObj;
};

async function updateCard(options) {
  const accessToken = await client.getAccessToken();
  const { cardInstanceId, cardData } = options;
  const data = {
    outTrackId: cardInstanceId,
    cardData: {
      cardParamMap: convertJSONValuesToString(cardData),
    },
    cardUpdateOptions: {
      updateCardDataByKey: true,
      updatePrivateDataByKey: true,
    },
  };
  const url = "https://api.dingtalk.com/v1.0/card/instances";
  try {
    const response = await axios.put(url, data, {
      headers: {
        "Content-Type": "application/json",
        Accept: "*/*",
        "x-acs-dingtalk-access-token": accessToken,
      },
    });
    if (response.data.success) {
      console.log("update card success: ", response.data);
    } else {
      console.error("update card failed: ", response.data);
    }
  } catch (error) {
    console.error("update card error: ", error);
  }
}
