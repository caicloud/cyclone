import { Table } from 'antd';
import PropTypes from 'prop-types';
import { FormatTime } from '@/lib/util';
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
      {
        title: intl.get('creationTime'),
        dataIndex: 'metadata.creationTimestamp',
        key: 'creationTime',
        render: value => FormatTime(value),
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
