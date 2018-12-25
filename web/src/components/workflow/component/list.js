import React from 'react';
import { Table, Button, Input } from 'antd';
import { inject, observer, PropTypes as MobxPropTypes } from 'mobx-react';
import PropTypes from 'prop-types';
import EllipsisMenu from '../../public/ellipsisMenu';
const Search = Input.Search;

@inject('workflow')
@observer
class List extends React.Component {
  static propTypes = {
    workflow: PropTypes.shape({
      workflowList: MobxPropTypes.observableArray,
      getWorkflowList: PropTypes.func,
    }),
  };
  componentDidMount() {
    this.props.workflow.getWorkflowList();
  }
  render() {
    const {
      workflow: { workflowList },
    } = this.props;
    const columns = [
      {
        title: intl.get('name'),
        dataIndex: 'name',
        key: 'name',
      },
      {
        title: intl.get('workflow.recentVersion'),
        dataIndex: 'recentVersion',
        key: 'recentVersion',
      },
      {
        title: intl.get('workflow.creator'),
        dataIndex: 'owner',
        key: 'owner',
      },
      {
        title: intl.get('creationTime'),
        dataIndex: 'creationTime',
        key: 'creationTime',
      },
      {
        title: intl.get('action'),
        dataIndex: 'action',
        render: () => <EllipsisMenu menuFunc={() => {}} />,
      },
    ];
    return (
      <div>
        <div className="head-bar">
          <Button type="primary">{intl.get('operation.add')}</Button>
          <Search
            placeholder="input search text"
            onSearch={() => {}}
            style={{ width: 200 }}
          />
        </div>
        <Table
          rowKey={row => row.id}
          columns={columns}
          dataSource={[...workflowList]}
        />
      </div>
    );
  }
}

export default List;
