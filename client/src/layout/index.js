import { Layout, Menu, Breadcrumb, Icon } from 'antd';
import React, { Component } from 'react';
import { NavLink } from 'react-router-dom';
import Routes from '../routes';
import { withRouter } from 'react-router-dom';
import PropTypes from 'prop-types';

const { SubMenu } = Menu;
const { Header, Content, Sider } = Layout;

class CoreLayout extends Component {
  static propTypes = {
    location: PropTypes.shape({
      pathname: PropTypes.string,
    }),
  };
  render() {
    const {
      location: { pathname },
    } = this.props;
    const MenuSelectKey = _.startsWith(pathname, '/pipeline')
      ? ['pipeline']
      : ['workspace'];
    return (
      <Layout>
        <Header className="header">
          <div className="logo" />
          <Menu
            theme="dark"
            mode="horizontal"
            defaultSelectedKeys={['2']}
            style={{ lineHeight: '64px' }}
          >
            <Menu.Item key="2">nav 2</Menu.Item>
            <Menu.Item key="3">nav 3</Menu.Item>
          </Menu>
        </Header>
        <Layout>
          <Sider width={200} style={{ background: '#fff' }}>
            <Menu
              mode="inline"
              defaultSelectedKeys={MenuSelectKey}
              defaultOpenKeys={['sub1']}
              style={{ height: '100%', borderRight: 0 }}
            >
              <SubMenu
                key="sub1"
                title={
                  <span>
                    <Icon type="user" />
                    subnav 1
                  </span>
                }
              >
                <Menu.Item key="workspace">
                  <NavLink to="/workspace" activeClassName="active">
                    流水线组
                  </NavLink>
                </Menu.Item>
                <Menu.Item key="pipeline">
                  <NavLink to="/pipeline" activeClassName="active">
                    流水线
                  </NavLink>
                </Menu.Item>
              </SubMenu>
            </Menu>
          </Sider>
          <Layout style={{ padding: '0 24px 24px' }}>
            <Breadcrumb style={{ margin: '16px 0' }}>
              <Breadcrumb.Item>Home</Breadcrumb.Item>
              <Breadcrumb.Item>List</Breadcrumb.Item>
              <Breadcrumb.Item>App</Breadcrumb.Item>
            </Breadcrumb>
            <Content
              style={{
                background: '#fff',
                padding: 24,
                margin: 0,
                minHeight: 280,
              }}
            >
              <Routes />
            </Content>
          </Layout>
        </Layout>
      </Layout>
    );
  }
}

export default withRouter(CoreLayout);
