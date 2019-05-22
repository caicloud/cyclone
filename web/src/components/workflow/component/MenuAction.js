import { Modal, Button } from 'antd';
import { inject, observer } from 'mobx-react';
import EllipsisMenu from '@/components/public/ellipsisMenu';
import RunWorkflow from '@/components/workflow/component/RunWorkflow';
import PropTypes from 'prop-types';

const Fragment = React.Fragment;
const confirm = Modal.confirm;

@inject('workflow')
@observer
class MenuAction extends React.Component {
  static propTypes = {
    workflow: PropTypes.object,
    projectName: PropTypes.string,
    workflowName: PropTypes.string,
    history: PropTypes.object,
    detail: PropTypes.boolean,
    visible: PropTypes.boolean,
  };

  state = {
    visible: false,
  };

  componentDidUpdate(preProps) {
    const { visible } = this.props;
    if (visible !== preProps.visible) {
      this.setState({ visible });
    }
  }

  removeWorkflow = () => {
    const { workflow, history, projectName, workflowName } = this.props;
    confirm({
      title: intl.get('confirmTip.remove', {
        resourceType: 'Workflow',
        name: workflowName,
      }),
      onOk() {
        workflow.deleteWorkflow(projectName, workflowName, () => {
          history.replace(`/workfow`);
        });
      },
    });
  };

  updateWorkflow = () => {
    const { history, projectName, workflowName } = this.props;
    history.push(`/workflow/${workflowName}/update?project=${projectName}`);
  };

  runWorkflow = () => {
    this.setState({ visible: true });
  };

  render() {
    const { detail, projectName, workflowName } = this.props;
    const { visible } = this.state;
    return (
      <Fragment>
        {detail && (
          <Button size="small" onClick={this.runWorkflow}>
            {intl.get('workflow.run')}
          </Button>
        )}
        <EllipsisMenu
          menuText={[
            intl.get('operation.modify'),
            intl.get('operation.delete'),
          ]}
          menuFunc={[
            () => {
              this.updateWorkflow();
            },
            () => {
              this.removeWorkflow();
            },
          ]}
        />
        <RunWorkflow
          visible={visible}
          projectName={projectName}
          workflowName={workflowName}
          handleModalClose={() => {
            this.setState({ visible: false });
          }}
        />
      </Fragment>
    );
  }
}
export default MenuAction;
