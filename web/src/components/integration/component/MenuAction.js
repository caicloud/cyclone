import { Modal } from 'antd';
import { inject, observer } from 'mobx-react';
import EllipsisMenu from '../../public/ellipsisMenu';
import PropTypes from 'prop-types';

const confirm = Modal.confirm;

@inject('integration')
@observer
class MenuAction extends React.Component {
  static propTypes = {
    integration: PropTypes.object,
    history: PropTypes.object,
    name: PropTypes.string,
    detail: PropTypes.bool,
  };
  removeIntegration = name => {
    const { integration, history } = this.props;
    confirm({
      title: `${intl.get('integration.confirm.tips')} ${name} ?`,
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
        menuText={[intl.get('operation.delete')]}
        menuFunc={[
          () => {
            this.removeIntegration(name);
          },
        ]}
      />
    );
  }
}
export default MenuAction;
