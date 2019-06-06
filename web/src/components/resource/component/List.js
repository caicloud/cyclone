import { Table, Button, Modal } from 'antd';
import PropTypes from 'prop-types';
import Resource from './Form';
import { inject, observer } from 'mobx-react';
import { getIntegrationName } from '@/lib/util';
import EllipsisMenu from '@/components/public/ellipsisMenu';

const Fragment = React.Fragment;
const confirm = Modal.confirm;

@inject('project')
@observer
class ResourceList extends React.Component {
  static propTypes = {
    projectName: PropTypes.string,
    project: PropTypes.object,
  };

  state = {
    visible: false,
    modifyData: null,
    update: false,
  };

  addResource = () => {
    this.setState({ visible: true });
  };

  componentDidMount() {
    const { projectName } = this.props;
    this.props.project.listProjectResources(projectName);
  }

  updateResource = (name, value) => {
    const integration = getIntegrationName(_.get(value, 'spec.parameters'));
    this.setState({
      visible: true,
      modifyData: { ...value, integration },
      update: true,
    });
  };

  removeResouece = name => {
    const {
      projectName,
      project: { deleteResource },
    } = this.props;
    confirm({
      title: intl.get('confirmTip.remove', {
        resourceType: 'Resource',
        name,
      }),
      onOk() {
        deleteResource(projectName, name);
      },
    });
  };

  render() {
    const { project, projectName } = this.props;
    const { visible, modifyData, update } = this.state;
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
      {
        title: intl.get('action'),
        dataIndex: 'metadata.name',
        key: 'action',
        align: 'right',
        render: (value, row) => (
          <EllipsisMenu
            menuText={[
              intl.get('operation.modify'),
              intl.get('operation.delete'),
            ]}
            menuFunc={[
              () => {
                this.updateResource(value, row);
              },
              () => {
                this.removeResouece(value);
              },
            ]}
          />
        ),
      },
    ];
    const resourceLen = list.length;
    return (
      <Fragment>
        <div className="head-bar">
          <Button type="primary" onClick={this.addResource}>
            {intl.get('operation.add')}
          </Button>
        </div>
        <Table
          loading={project.loadingResource}
          columns={columns}
          rowKey={record => record.metadata.name}
          dataSource={list}
        />
        {visible && (
          <Resource
            handleModalClose={() => {
              this.setState({
                visible: false,
                modifyData: null,
                update: false,
              });
            }}
            visible={visible}
            update={update}
            projectName={projectName}
            resourceLen={resourceLen}
            modifyData={modifyData}
          />
        )}
      </Fragment>
    );
  }
}

export default ResourceList;
