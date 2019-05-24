import { Modal } from 'antd';
import { inject, observer } from 'mobx-react';
import EllipsisMenu from '@/components/public/ellipsisMenu';
import PropTypes from 'prop-types';

const confirm = Modal.confirm;

@inject('integration')
@observer
class MenuAction extends React.Component {
  static propTypes = {
    integration: PropTypes.object,
    history: PropTypes.object,
    name: PropTypes.string,
  };
  removeIntegration = name => {
    const { integration, history } = this.props;
    confirm({
      title: `${intl.get('integration.confirm.tips')} ${name} ?`,
      onOk() {
        integration.deleteIntegration(name, () => {
          history.replace('/integration');
          integration.getIntegrationList();
        });
      },
    });
  };

  updateIntegration = name => {
    const { history } = this.props;
    history.push(`/integration/${name}/update`);
  };

  render() {
    const { name } = this.props;
    return (
      <EllipsisMenu
        menuText={[intl.get('operation.update'), intl.get('operation.delete')]}
        menuFunc={[
          () => {
            this.updateIntegration(name);
          },
          () => {
            this.removeIntegration(name);
          },
        ]}
      />
    );
  }
}
export default MenuAction;
