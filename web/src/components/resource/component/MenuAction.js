import { Modal } from 'antd';
import { inject, observer } from 'mobx-react';
import EllipsisMenu from '@/components/public/ellipsisMenu';
import PropTypes from 'prop-types';

const confirm = Modal.confirm;

@inject('resource')
@observer
class MenuAction extends React.Component {
  static propTypes = {
    project: PropTypes.object,
    type: PropTypes.string,
    history: PropTypes.object,
    resource: PropTypes.object,
  };

  removeResourceType = type => {
    const { resource, history } = this.props;
    confirm({
      title: intl.get('confirmTip.removeResourceType', { resourceType: type }),
      onOk() {
        resource.deleteResourceType(type, () => {
          history.replace(`/resource`);
          resource.listResourceTypes();
        });
      },
    });
  };

  updateResourceType = type => {
    const { history } = this.props;
    history.push(`/resource/${type}/update`);
  };

  render() {
    const { type } = this.props;
    return (
      <EllipsisMenu
        menuText={[intl.get('operation.modify'), intl.get('operation.delete')]}
        menuFunc={[
          () => {
            this.updateResourceType(type);
          },
          () => {
            this.removeResourceType(type);
          },
        ]}
      />
    );
  }
}
export default MenuAction;
