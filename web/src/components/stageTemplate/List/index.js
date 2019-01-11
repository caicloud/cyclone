import { inject, observer } from 'mobx-react';
import PropTypes from 'prop-types';
import React from 'react';
import { Layout } from 'antd';

import List from './List';

const { Content } = Layout;

@inject('stageTemplate')
@observer
class StageTemplate extends React.Component {
  componentDidMount() {
    this.props.stageTemplate.getTemplateList();
  }
  render() {
    const {
      stageTemplate: { templateList = [] },
    } = this.props;
    return (
      <Layout style={{ background: 'transparent' }}>
        {/* TODO: add button for creating */}
        {/* <Header style={{ background: 'transparent', padding: 0 }} /> */}
        {/* TODO: add tag-filter */}
        {/* <Sider
          style={{
            background: 'transparent',
            overflow: 'auto',
          }}
        >
        </Sider> */}
        <Content>
          <List list={templateList.concat(templateList)} />
        </Content>
      </Layout>
    );
  }
}

StageTemplate.propTypes = {
  stageTemplate: PropTypes.object,
};

export default StageTemplate;
