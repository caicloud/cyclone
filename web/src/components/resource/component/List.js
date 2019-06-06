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
    showDetail: false,
  };

  addResource = () => {
    this.setState({ visible: true });
  };

  componentDidMount() {
    const { projectName } = this.props;
    this.props.project.listProjectResources(projectName);
  }

  updateResource = (name, value, detail) => {
    const integration = getIntegrationName(_.get(value, 'spec.parameters'));
    let state = {
      modifyData: { ...value, integration },
    };
    if (detail) {
      state.detailVisible = true;
    } else {
      state.update = true;
      state.visible = true;
    }
    this.setState(state);
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
    const { visible, modifyData, update, detailVisible } = this.state;
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
          onRow={row => {
            return {
              onClick: () => {
                this.updateResource(_.get(row, 'metadata.name'), row, 'detail');
              },
            };
          }}
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
        {detailVisible && (
          <Resource
            handleModalClose={() => {
              this.setState({
                detailVisible: false,
                modifyData: null,
              });
            }}
            title={intl.get('resource.detail')}
            visible={detailVisible}
            projectName={projectName}
            resourceLen={resourceLen}
            modifyData={modifyData}
            readOnly={true}
          />
        )}
      </Fragment>
    );
  }
}

export default ResourceList;
