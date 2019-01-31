import { Layout, Menu, Icon } from 'antd';
import React, { Component } from 'react';
import { NavLink } from 'react-router-dom';
import Routes from '../routes';
import { withRouter } from 'react-router-dom';
import PropTypes from 'prop-types';
import BreadCrumb from './Breadcrumb';

const { Header, Content, Sider } = Layout;

const siderWidth = 200;

class CoreLayout extends Component {
  static propTypes = {
    location: PropTypes.shape({
      pathname: PropTypes.string,
    }),
  };

  state = {
    collapsed: false,
  };

  onCollapse = collapsed => {
    this.setState({ collapsed });
  };

  render() {
    const { location } = this.props;
    const pathSnippets = location.pathname.split('/').filter(i => i);
    const selectNav = pathSnippets[0] ? pathSnippets[0] : '/overview';
    return (
      <Layout style={{ minHeight: '100%' }}>
        <Header className="cyclone-layout-header">
          <div className="cyclone-logo">CYCLONE</div>
          <Menu theme="dark" mode="horizontal" style={{ lineHeight: '64px' }}>
            <Menu.Item key="3">
              <Icon type="user" />
            </Menu.Item>
          </Menu>
        </Header>
        <Layout style={{ marginTop: 64 }}>
          <Sider
            width={siderWidth}
            style={{
              background: '#fff',
              overflow: 'auto',
              height: '100vh',
              position: 'fixed',
              left: 0,
            }}
            collapsed={this.state.collapsed}
            onCollapse={this.onCollapse}
          >
            <Menu
              mode="inline"
              defaultSelectedKeys={[selectNav]}
              style={{ height: '100%', borderRight: 0 }}
            >
              <Menu.Item key="overview">
                <NavLink to="/overview" activeClassName="active">
                  <Icon type="home" />
                  <span>{intl.get('sideNav.overview')}</span>
                </NavLink>
              </Menu.Item>
              <Menu.Item key="project">
                <NavLink to="/project" activeClassName="active">
                  <Icon type="project" />
                  <span>{intl.get('sideNav.project')}</span>
                </NavLink>
              </Menu.Item>
              <Menu.Item key="stageTemplate">
                <NavLink to="/stageTemplate" activeClassName="active">
                  <Icon type="profile" />
                  <span>{intl.get('sideNav.stageTemplate')}</span>
                </NavLink>
              </Menu.Item>
              <Menu.Item key="workflow">
                <NavLink to="/workflow" activeClassName="active">
                  <Icon type="share-alt" />
                  <span>{intl.get('sideNav.workflow')}</span>
                </NavLink>
              </Menu.Item>
              <Menu.Item key="integration">
                <NavLink to="/integration" activeClassName="active">
                  <Icon type="sliders" />
                  <span>{intl.get('sideNav.integration')}</span>
                </NavLink>
              </Menu.Item>
              {/* TODO: manage and setting */}
              {/* <SubMenu
                key="manage"
                title={
                  <span>
                    <Icon type="team" />
                    管理中心
                  </span>
                }
              >
                <Menu.Item key="tenant">租户</Menu.Item>
                <Menu.Item key="user">
                  <NavLink activeClassName="active">
                    用户
                  </NavLink>
                </Menu.Item>
              </SubMenu>
              <SubMenu
                key="setting"
                title={
                  <span>
                    <Icon type="setting" />
                    配置中心
                  </span>
                }
              >
                <Menu.Item>
                  <span>设置</span>
                </Menu.Item>
              </SubMenu> */}
            </Menu>
          </Sider>
          <Layout
            style={{ padding: '12px 24px 24px 24px', marginLeft: siderWidth }}
          >
            <BreadCrumb location={location} />
            <Content
              style={{
                background: '#fff',
                padding: 24,
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
