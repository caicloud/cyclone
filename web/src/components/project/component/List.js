import React from 'react';
import { Table, Button } from 'antd';
import { inject, observer } from 'mobx-react';
import PropTypes from 'prop-types';
import EllipsisMenu from '../../public/ellipsisMenu';

@inject('project')
@observer
class List extends React.Component {
  /**
   * TODO: list project
   * submit crete form action
   */
  static propTypes = {
    match: PropTypes.object,
    history: PropTypes.object,
    project: PropTypes.object,
  };
  componentDidMount() {
    const { project } = this.props;
    project.listProjects();
  }
  saveFormRef = formRef => {
    this.formRef = formRef;
  };

  showModal = () => {
    const { match } = this.props;
    this.props.history.push(`${match.path}/add`);
  };

  handleCreate = () => {
    const form = this.formRef.props.form;
    form.validateFields((err, values) => {
      if (err) {
        return;
      }
      form.resetFields();
    });
  };

  // TODO(qme): finish remove projct
  // removeProject = name => {
  //   confirm({
  //     title: `Do you Want to delete project ${name} ?`,
  //     onOk() {
  //       this.props.project.deleteProject(name);
  //     },
  //   });
  // };
  render() {
    const { match, project } = this.props;
    const columns = [
      {
        title: intl.get('name'),
        dataIndex: 'metadata.name',
        key: 'name',
      },
      {
        title: intl.get('creationTime'),
        dataIndex: 'metadata.creationTimestamp',
        key: 'creationTime', // TODO(qme): transform time
      },
      {
        title: intl.get('project.workflowCount'),
        dataIndex: 'workflowCount',
        key: 'workflowCount',
      },
      {
        title: intl.get('action'),
        dataIndex: 'metadata.name',
        key: 'action',
        render: value => <EllipsisMenu menuFunc={() => {}} />,
      },
    ];
    return (
      <div>
        <div className="head-bar">
          <Button type="primary" onClick={this.showModal}>
            {intl.get('operation.add')}
          </Button>
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
          dataSource={_.get(project, 'projectList.items', [])}
        />
      </div>
    );
  }
}

export default List;
