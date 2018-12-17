import React from 'react';
import { Table } from 'antd';
import { inject, observer } from 'mobx-react';
import PropTypes from 'prop-types';

@inject('pipeline')
@observer
class List extends React.Component {
  static propTypes = {
    pipeline: PropTypes.object,
  };
  componentDidMount() {
    this.props.pipeline.getPipelineList();
  }
  render() {
    const { pipeline } = this.props;
    const columns = [
      {
        title: 'Name',
        dataIndex: 'name',
        key: 'name',
      },
      { title: 'Alias', dataIndex: 'alias', key: 'alias' },
    ];
    return <Table columns={columns} dataSource={pipeline.pipelineList} />;
  }
}

export default List;
