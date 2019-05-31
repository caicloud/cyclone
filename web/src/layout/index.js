import { Layout, Menu, Icon, Popover, Tooltip } from 'antd';
import React, { Component } from 'react';
import { NavLink } from 'react-router-dom';
import Routes from '../routes';
import { withRouter } from 'react-router-dom';
import PropTypes from 'prop-types';
import BreadCrumb from './Breadcrumb';
import CycloneLogo from '@/images/logo.png';

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

  onLocaleChange = item => {
    if (item.key !== localStorage.getItem('lang')) {
      localStorage.setItem('lang', item.key);
      window.location.reload();
    }
  };

  render() {
    const { location } = this.props;
    const pathSnippets = location.pathname.split('/').filter(i => i);
    const selectNav = pathSnippets[0] ? pathSnippets[0] : '/overview';
    const lang = localStorage.getItem('lang') || 'en-US';

    const languages = (
      <Menu style={{ borderRight: 'none' }}>
        <Menu.Item
          key="zh-CN"
          disabled={lang === 'zh-CN'}
          onClick={this.onLocaleChange}
        >
          中文
        </Menu.Item>
        <Menu.Item
          key="en-US"
          disabled={lang === 'en-US'}
          onClick={this.onLocaleChange}
        >
          ENGLISH
        </Menu.Item>
      </Menu>
    );

    return (
      <Layout style={{ minHeight: '100%' }}>
        <Header className="cyclone-layout-header">
          <div className="cyclone-logo">
            <img src={CycloneLogo} alt="cyclone logo" />
            CYCLONE
          </div>
          <div>
            <Icon className="headbar-icon" type="user" />
            <Popover
              placement="bottomRight"
              trigger="click"
              content={languages}
            >
              <Icon
                className="headbar-icon"
                type="global"
                onClick={this.onLangClick}
              />
            </Popover>
            <Tooltip
              placement="bottomRight"
              title={intl.get('tips.officialWebsite')}
            >
              <a
                href="https://cyclone.dev"
                target="_blank"
                rel="noopener noreferrer"
              >
                <Icon className="headbar-icon" type="book" />
              </a>
            </Tooltip>
          </div>
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
              <Menu.Item key="resource">
                <NavLink to="/resource" activeClassName="active">
                  <Icon type="cloud" />
                  <span>{intl.get('sideNav.resource')}</span>
                </NavLink>
              </Menu.Item>
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
