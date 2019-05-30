import { Formik } from 'formik';
import { Modal, notification } from 'antd';
import BindResource from './BindResource';
import PropTypes from 'prop-types';
import { inject, observer } from 'mobx-react';
import { resourceParametersField } from '@/lib/const';

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
    const { visible } = props;
    this.state = {
      visible,
    };
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

  getInitialValues = () => {
    const { type, modifyData } = this.props;
    let data = {
      name: '',
      path: '',
      type: 'Git',
      spec: {
        parameters:
          type === 'inputs'
            ? resourceParametersField['SCM']
            : resourceParametersField['DockerRegistry'],
      },
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
      ..._.pick(value, ['spec.parameters', 'spec.type']),
    };
    const modifyResource = !_.isEmpty(modifyData);
    if (update) {
      if (modifyResource) {
        updateResource(
          project,
          _.get(value, 'metadata.name'),
          resourceObj,
          () => {
            SetReasourceValue(value, modifyResource);
            notification.success({
              message: '更新 Resource',
              duration: 2,
            });
          }
        );
      } else {
        createResource(project, resourceObj, () => {
          SetReasourceValue(value, modifyResource);
        });
      }
    } else {
      SetReasourceValue(value, modifyResource);
    }
    this.setState({ visible: false });
    handleModalClose(false);
  };
  render() {
    const { type, update, modifyData } = this.props;
    const { visible } = this.state;
    return (
      <Formik
        initialValues={this.getInitialValues()}
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
