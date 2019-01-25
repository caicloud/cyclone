import React from 'react';
import { Table, Button, Modal } from 'antd';
import { inject, observer } from 'mobx-react';
import PropTypes from 'prop-types';
import EllipsisMenu from '../../public/ellipsisMenu';

const confirm = Modal.confirm;

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

  removeProject = name => {
    const { project } = this.props;
    confirm({
      title: `Do you Want to delete project ${name} ?`,
      onOk() {
        project.deleteProject(name);
      },
    });
  };
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
        render: value => (
          <EllipsisMenu
            menuFunc={() => {
              this.removeProject(value);
            }}
          />
        ),
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
          rowKey={record => record.name}
          onRow={record => {
            return {
              onClick: () => {
                this.props.history.push(
                  `${match.path}/${record.metadata.name}`
                );
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
