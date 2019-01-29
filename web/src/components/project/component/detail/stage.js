import { Table } from 'antd';
import PropTypes from 'prop-types';
import { inject, observer } from 'mobx-react';

@inject('project')
@observer
class StageList extends React.Component {
  static propTypes = {
    projectName: PropTypes.string,
    project: PropTypes.object,
  };
  componentDidMount() {
    const { projectName } = this.props;
    this.props.project.listProjectStages(projectName);
  }

  render() {
    const { project } = this.props;
    const list = _.get(project, ['stageList', 'items'], []);
    const columns = [
      {
        title: intl.get('name'),
        dataIndex: 'metadata.name',
        key: 'name',
      },
    ];

    return (
      <Table
        loading={project.loadingStage}
        columns={columns}
        rowKey={record => record.metadata.name}
        dataSource={list}
      />
    );
  }
}

export default StageList;
