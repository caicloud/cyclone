import { Formik } from 'formik';
import { Modal, notification, Spin } from 'antd';
import BindResource from './BindResource';
import PropTypes from 'prop-types';
import { inject, observer } from 'mobx-react';

@inject('resource')
@observer
class ResourceFrom extends React.Component {
  static propTypes = {
    SetReasourceValue: PropTypes.func,
    type: PropTypes.oneOf(['inputs', 'ouputs']),
    visible: PropTypes.bool,
    modifyData: PropTypes.object,
    handleModalClose: PropTypes.func,
    update: PropTypes.bool,
    project: PropTypes.string,
    resource: PropTypes.object,
  };

  static defaultProps = {
    type: 'inputs',
  };

  constructor(props) {
    super(props);
    const {
      visible,
      type,
      resource: { listResourceTypes },
    } = props;
    this.state = {
      visible,
      loading: true,
    };
    listResourceTypes(
      { operation: type === 'inputs' ? 'pull' : 'push' },
      () => {
        this.setState({ loading: false });
      }
    );
  }

  componentDidUpdate(preProps) {
    const { visible } = this.props;
    if (visible !== preProps.visible) {
      this.setState({ visible });
    }
  }

  closeModal = () => {
    const { handleModalClose } = this.props;
    this.setState({ visible: false });
    handleModalClose && handleModalClose(false);
  };

  getInitialValues = list => {
    const { modifyData } = this.props;
    const item = _.get(modifyData, 'type')
      ? _.find(list, o => _.get(o, 'spec.type') === _.get(modifyData, 'type'))
      : list[0];
    let data = {
      name: '',
      path: '',
      type: _.get(item, 'spec.type'),
      ..._.pick(item, ['spec.parameters']),
    };
    if (!_.isEmpty(modifyData)) {
      data = _.merge(data, modifyData);
    }
    return data;
  };

  submitResource = value => {
    const {
      SetReasourceValue,
      handleModalClose,
      update,
      project,
      modifyData,
      resource: { createResource, updateResource },
    } = this.props;
    const resourceObj = {
      metadata: { name: _.get(value, 'name') },
      ..._.pick(value, ['spec.parameters']),
    };
    const modifyResource = !_.isEmpty(modifyData);
    if (update) {
      resourceObj.spec.type = _.get(value, 'type');
      if (!modifyResource) {
        createResource(project, resourceObj, () => {
          SetReasourceValue(value, modifyResource);
        });
      } else if (!_.isEqual(value, modifyData)) {
        updateResource(project, _.get(value, 'name'), resourceObj, () => {
          SetReasourceValue(value, modifyResource);
          notification.success({
            message: Intl.get('notification.updateResource'),
            duration: 2,
          });
        });
      }
    } else {
      SetReasourceValue(value, modifyResource);
    }
    this.setState({ visible: false });
    handleModalClose(false);
  };
  render() {
    const {
      type,
      update,
      modifyData,
      resource: { resourceTypeList },
    } = this.props;
    const { visible, loading } = this.state;
    if (loading) {
      return <Spin />;
    }
    return (
      <Formik
        initialValues={this.getInitialValues(
          _.get(resourceTypeList, 'items', [])
        )}
        onSubmit={this.submitResource}
        render={props => (
          <Modal
            visible={visible}
            onCancel={this.closeModal}
            onOk={props.handleSubmit}
          >
            <BindResource
              {...props}
              type={type}
              update={update && !_.isEmpty(modifyData)}
            />
          </Modal>
        )}
      />
    );
  }
}

export default ResourceFrom;
