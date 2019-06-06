import { Formik } from 'formik';
import { Modal, Spin } from 'antd';
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
    resource: PropTypes.object,
    projectName: PropTypes.string,
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
    };
    if (!_.isEmpty(modifyData)) {
      data = _.merge(data, modifyData);
    }
    return data;
  };

  submitResource = value => {
    const { modifyData, SetReasourceValue, handleModalClose } = this.props;
    const modify = !!modifyData;
    SetReasourceValue(value, modify);
    this.setState({ visible: false });
    handleModalClose && handleModalClose(false);
  };
  render() {
    const {
      type,
      update,
      modifyData,
      projectName,
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
              projectName={projectName}
            />
          </Modal>
        )}
      />
    );
  }
}

export default ResourceFrom;
