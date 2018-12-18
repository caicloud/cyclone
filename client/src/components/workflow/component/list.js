import React from 'react';
import { Table, Button, Input } from 'antd';
import { inject, observer } from 'mobx-react';
import PropTypes from 'prop-types';
const Search = Input.Search;

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
        title: '名称',
        dataIndex: 'name',
        key: 'name',
      },
      {
        title: '最近版本',
        dataIndex: 'name',
        key: 'name',
      },
      {
        title: '创建者',
        dataIndex: 'name',
        key: 'name',
      },
      {
        title: '创建时间',
        dataIndex: 'name',
        key: 'name',
      },
      {
        title: '操作',
        dataIndex: 'name',
        key: 'name',
        render: () => {
          return (
            <div>
              <Button>执行</Button>
            </div>
          );
        },
      },
    ];
    return (
      <div>
        <div className="head-bar">
          <Button type="primary">新增</Button>
          <Search
            placeholder="input search text"
            onSearch={value => console.log(value)}
            style={{ width: 200 }}
          />
        </div>
        <Table columns={columns} dataSource={pipeline.pipelineList} />
      </div>
    );
  }
}

export default List;
