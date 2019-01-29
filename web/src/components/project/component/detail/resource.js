import { Table } from 'antd';
import PropTypes from 'prop-types';
import { inject, observer } from 'mobx-react';

@inject('project')
@observer
class ResourceList extends React.Component {
  static propTypes = {
    projectName: PropTypes.string,
    project: PropTypes.object,
  };
  componentDidMount() {
    const { projectName } = this.props;
    this.props.project.listProjectResources(projectName);
  }

  render() {
    const { project } = this.props;
    const list = _.get(project, ['resourceList', 'items'], []);
    const columns = [
      {
        title: intl.get('name'),
        dataIndex: 'metadata.name',
        key: 'name',
      },
      {
        title: intl.get('type'),
        dataIndex: 'spec.type',
        key: 'type',
      },
    ];

    return (
      <Table
        loading={project.loadingResource}
        columns={columns}
        rowKey={record => record.metadata.name}
        dataSource={list}
      />
    );
  }
}

export default ResourceList;
