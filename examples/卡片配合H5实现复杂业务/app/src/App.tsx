import React, { useState, useEffect } from "react";
import logo from "./logo.svg";
import { Input, Table, Form, Typography, Rate, Checkbox, Button } from "antd";
import "./App.css";

const { TextArea } = Input;
const { Title, Paragraph } = Typography;

const columns = [
  {
    title: "应用名",
    dataIndex: "appName",
    key: "appName",
  },
  {
    title: "点击次数",
    dataIndex: "pv",
    key: "pv",
  },
  {
    title: "点击人数",
    dataIndex: "uv",
    key: "uv",
  },
];

const dataSource = [
  {
    key: "1",
    appName: "考勤打卡",
    pv: 433,
    uv: 324,
    children: [
      {
        key: "1-1",
        appName: "今天",
        pv: 200,
        uv: 144,
      },
      {
        key: "1-2",
        appName: "昨天",
        pv: 133,
        uv: 100,
      },
      {
        key: "1-3",
        appName: "前天",
        pv: 100,
        uv: 80,
      },
    ],
  },
  {
    key: "2",
    appName: "智能人事",
    pv: 354,
    uv: 350,
    children: [
      {
        key: "2-1",
        appName: "今天",
        pv: 187,
        uv: 185,
      },
      {
        key: "2-2",
        appName: "昨天",
        pv: 99,
        uv: 98,
      },
      {
        key: "2-3",
        appName: "昨天",
        pv: 68,
        uv: 67,
      },
    ],
  },
  {
    key: "3",
    appName: "日志",
    pv: 322,
    uv: 189,
    children: [
      {
        key: "3-1",
        appName: "今天",
        pv: 166,
        uv: 95,
      },
      {
        key: "3-2",
        appName: "昨天",
        pv: 72,
        uv: 40,
      },
      {
        key: "3-3",
        appName: "昨天",
        pv: 84,
        uv: 54,
      },
    ],
  },
];

type FieldType = {
  stars: number;
  tags?: string[];
  suggestion?: string;
};

function App() {
  const [queryParams, setQueryParams] = useState<Record<string, string>>({});
  const [submitted, setSubmitted] = useState(false);
  const [showFeedback, setShowFeedback] = useState(false);

  useEffect(() => {
    const searchParams = new URLSearchParams(window.location.search);
    const params: typeof queryParams = {};
    for (let [key, value] of searchParams) {
      params[key] = value;
    }
    setQueryParams(params);
  }, []);

  const onFinish = async (values: FieldType) => {
    console.log("submit values: ", values);
    // 调用接口提交表单数据，并在服务端调用更新卡片数据接口把表单数据更新到卡片上
    setSubmitted(true);
  };

  // TODO: 添加钉钉免登鉴权

  if (queryParams.page === "detail" && queryParams.id) {
    return (
      <div style={{ width: "100%", maxWidth: "600px", margin: "0 auto" }}>
        <Typography>
          <Title style={{ textAlign: "center" }}>
            Table id: {queryParams.id}
          </Title>
          <Paragraph>
            <Table columns={columns} dataSource={dataSource} />
          </Paragraph>
        </Typography>
      </div>
    );
  }

  if (queryParams.page === "evaluate" && queryParams.id) {
    return (
      <div
        style={{
          width: "100%",
          maxWidth: "600px",
          margin: "0 auto",
          padding: "12px",
        }}
      >
        <Typography>
          <Title style={{ textAlign: "center" }}>
            Form id: {queryParams.id}
          </Title>
          <Form
            name="evaluate"
            onFinish={onFinish}
            onValuesChange={(changedValues: FieldType) => {
              if (changedValues.stars != null) {
                setShowFeedback(changedValues.stars < 4);
              }
            }}
          >
            <Form.Item
              label="评价"
              name="stars"
              rules={[{ required: true, message: "请选择评价" }]}
            >
              <Rate allowHalf />
            </Form.Item>
            {showFeedback && (
              <>
                <Form.Item label="反馈" name="tags">
                  <Checkbox.Group
                    options={[
                      "评价 A",
                      "评价 B",
                      "评价 C",
                      "评价 D",
                      "评价 E",
                    ].map((x) => ({ label: x, value: x }))}
                  />
                </Form.Item>
                <Form.Item label="建议" name="suggestion">
                  <TextArea autoSize={{ minRows: 4, maxRows: 10 }} />
                </Form.Item>
              </>
            )}
            <Form.Item>
              <Button type="primary" htmlType="submit" disabled={submitted}>
                {submitted ? "已提交" : "提交"}
              </Button>
            </Form.Item>
          </Form>
        </Typography>
      </div>
    );
  }

  return (
    <div className="App">
      <header className="App-header">
        <img src={logo} className="App-logo" alt="logo" />
        <p>
          Edit <code>src/App.tsx</code> and save to reload.
        </p>
        <p>{JSON.stringify(queryParams)}</p>
        <a
          className="App-link"
          href="https://reactjs.org"
          target="_blank"
          rel="noopener noreferrer"
        >
          Learn React
        </a>
      </header>
    </div>
  );
}

export default App;
