import React, {useState} from 'react';
import './App.css';
import 'antd/dist/antd.css';
import MyGallery from "./pages/gallery";
import TreeFolder from "./pages/treeFolder";
import {Layout, Menu} from 'antd';
import {LogoutOutlined} from "@ant-design/icons";


function App() {
    const { Sider,Content } = Layout;

    const [collapsed,setCollapsed] = useState(false)

    const toggleCollapsed = () => {
        setCollapsed(!collapsed);
    };
    const [urlFolder,setUrlFolder] = useState('')
  return (
      <Layout>
              <Sider collapsible collapsed={collapsed} onCollapse={toggleCollapsed} width={300}>
                  <div className="logo" />
                  <Content style={{height:100+'%'}}>
                      <Menu theme={"dark"}>
                          <Menu.Item>
                              <LogoutOutlined /><span>Mettre à jour</span>
                          </Menu.Item>
                      </Menu>
                      {!collapsed ? <TreeFolder setUrlFolder={setUrlFolder}/>:<></>}

                  </Content>
              </Sider>
          <Layout>
              <MyGallery urlFolder={urlFolder}/>
          </Layout>
      </Layout>
  );
}

export default App;
