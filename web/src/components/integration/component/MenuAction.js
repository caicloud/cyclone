import { Modal } from 'antd';
import { inject, observer, PropTypes } from 'mobx-react';
import EllipsisMenu from '../../public/ellipsisMenu';

const confirm = Modal.confirm;

@inject('integration')
@observer
class MenuAction extends React.Component {
  static propTypes = {
    integration: PropTypes.observableObject,
    history: PropTypes.object,
    name: PropTypes.string,
    detail: PropTypes.bool,
  };
  removeProject = name => {
    const { integration, history } = this.props;
    confirm({
      title: `Do you Want to delete project ${name} ?`,
      onOk() {
        integration.deleteIntegration(name, () => {
          history.replace('/integration');
        });
      },
    });
  };

  render() {
    const { name } = this.props;
    return (
      <EllipsisMenu
        menuText={[intl.get('operation.modify'), intl.get('operation.delete')]}
        menuFunc={[
          () => {
            this.updateProject(name);
          },
          () => {
            this.removeProject(name);
          },
        ]}
      />
    );
  }
}
export default MenuAction;
