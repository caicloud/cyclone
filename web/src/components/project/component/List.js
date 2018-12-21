import React from 'react';
import { Table, Button } from 'antd';
import { inject, observer } from 'mobx-react';
import PropTypes from 'prop-types';
import EllipsisMenu from '../../public/ellipsisMenu';
import CreateProjectForm from './AddProject';

@inject('workflow')
@observer
class List extends React.Component {
  /**
   * TODO: list project
   * submit crete form action
   */
  static propTypes = {
    match: PropTypes.object,
    history: PropTypes.object,
  };
  state = {
    visible: false,
  };

  saveFormRef = formRef => {
    this.formRef = formRef;
  };

  showModal = () => {
    this.setState({ visible: true });
  };

  handleCancel = () => {
    this.setState({ visible: false });
  };

  handleCreate = () => {
    const form = this.formRef.props.form;
    form.validateFields((err, values) => {
      if (err) {
        return;
      }
      form.resetFields();
      this.setState({ visible: false });
    });
  };
  render() {
    const { visible } = this.state;
    const { match } = this.props;
    const columns = [
      {
        title: intl.get('name'),
        dataIndex: 'name',
        key: 'name',
      },
      {
        title: intl.get('creationTime'),
        dataIndex: 'creationTime',
        key: 'creationTime',
      },
      {
        title: intl.get('project.workflowCount'),
        dataIndex: 'workflowCount',
        key: 'workflowCount',
      },
      {
        title: intl.get('action'),
        dataIndex: 'action',
        key: 'action',
        render: () => <EllipsisMenu menuFunc={() => {}} />,
      },
    ];
    return (
      <div>
        <div className="head-bar">
          <Button type="primary" onClick={this.showModal}>
            {intl.get('operation.add')}
          </Button>
          <CreateProjectForm
            wrappedComponentRef={this.saveFormRef}
            visible={visible}
            onCancel={this.handleCancel}
            onCreate={this.handleCreate}
          />
        </div>
        <Table
          columns={columns}
          rowKey={record => record.id}
          onRow={record => {
            return {
              onClick: () => {
                this.props.history.push(`${match.path}/${record.id}`);
              },
            };
          }}
          dataSource={[
            {
              name: '项目1',
              id: 'proejct1',
              creationTime: '2018-12-26 09:09',
              workflowCount: '2',
            },
          ]}
        />
      </div>
    );
  }
}

export default List;
