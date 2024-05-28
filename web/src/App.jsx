import { useState, useEffect } from "react";
import "./App.css";
import { Button, Card, Tree, Table, Input } from "antd";
import axios from 'axios';

const dig = (path = "0", level = 8) => {
  const list = [];
  for (let i = 0; i < 2; i += 1) {
    const key = `${path}-${i}`;
    const treeNode = {
      title: key,
      key,
    };
    if (level > 0) {
      treeNode.children = dig(key, level - 1);
    }
    list.push(treeNode);
  }
  return list;
};



const App = () => {
  const [dataList, setDataList] = useState([]);
  const treeData = dig();
  const [expandedKeys, setExpandedKeys] = useState([]);
  const [searchValue, setSearchValue] = useState("");
  const [autoExpandParent, setAutoExpandParent] = useState(true);
  const onExpand = (newExpandedKeys) => {
    setExpandedKeys(newExpandedKeys);
    setAutoExpandParent(false);
  };

  useEffect(()=> {
    axios.get('/api/get-childrencmds?id=2&depth=1').then(res => {
      setDataList(res.data);
    })
  }, [])
  // const onChange = (e) => {
  //   const { value } = e.target;
  //   const newExpandedKeys = dataList
  //     .map((item) => {
  //       if (item.title.indexOf(value) > -1) {
  //         return getParentKey(item.key, defaultData);
  //       }
  //       return null;
  //     })
  //     .filter((item, i, self) => item && self.indexOf(item) === i);
  //   setExpandedKeys(newExpandedKeys);
  //   setSearchValue(value);
  //   setAutoExpandParent(true);
  // };
  
  const updateTreeData = (list, key, children) =>{
  return list.map((node) => {
    console.log(node)
    if (node.id === key) {
      return {
        ...node,
        children,
      };
    }
    if (node.children) {
      return {
        ...node,
        children: updateTreeData(node.children, key, children),
      };
    }
    return node;
  })};
  
  const onLoadData = ({key,children})=> 
    axios.get(`/api/get-childrencmds?id=${key}&depth=1`).then((res)=> {
      const newTreeData = updateTreeData(dataList, key, res.data);
      console.log(newTreeData)
      setDataList(newTreeData);
    })
  const getKey = () => {
    const _data = [];
    const nTree = (data) =>
      data.forEach((item) => {
        _data.push(item.cmd);
        if (item.children) {
          nTree(item.children);
        }
      });
    nTree(dataList);
    return _data;
  };
  const filterTreeNode = (treeNode) => {
    if (!searchValue) return false;  // variableKeyWord：关键字 自己写搜索框逻辑
    console.log(treeNode, searchValue)
    return treeNode?.title?.indexOf(searchValue) > -1;
  }

  return (
    <>
      <Card style={{ marginBottom: "20px" }}>
        <div
          style={{
            width: "100%",
            display: "flex",
            justifyContent: "end",
            gap: "10px",
          }}
        >
          <Button>省略展示</Button>
          <Button onClick={() => setExpandedKeys(getKey())}>全部展开</Button>
          <Button onClick={() => setExpandedKeys([])}>全部折叠</Button>
        </div>
      </Card>
      <Card>
        <Input.Search
          style={{
            marginBottom: 8,
          }}
          placeholder="Search"
          onChange={(e) => setSearchValue(e.target.value)}
        />
        <Tree
          onExpand={onExpand}
          expandedKeys={expandedKeys}
          autoExpandParent={autoExpandParent}
          treeData={dataList}
          style={{ width: "100%" }}
          height={700}
          filterTreeNode={filterTreeNode}
          fieldNames={{
            title: 'cmd',
             key: 'id',
          }}
          loadData={onLoadData}
        ></Tree>
      </Card>

      {/* <Table
        columns={[{ title: "Name", dataIndex: "title" }]}
        dataSource={treeData}
      ></Table> */}
    </>
  );
};

export default App;
