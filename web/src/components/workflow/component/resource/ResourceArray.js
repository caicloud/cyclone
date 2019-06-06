import { Button, Row, Col, Form } from 'antd';
import { FieldArray, FastField } from 'formik';
import { drawerFormItemLayout } from '@/lib/const';
import Resource from './Form';
import PropTypes from 'prop-types';
import style from '@/components/workflow/component/index.module.less';

const FormItem = Form.Item;
const Fragment = React.Fragment;

class ResourceArray extends React.Component {
  static propTypes = {
    resourcesField: PropTypes.string,
    type: PropTypes.oneOf(['inputs', 'outputs']),
    update: PropTypes.bool,
    projectName: PropTypes.string,
    resources: PropTypes.array,
  };

  state = {
    visible: false,
  };

  getIntegrationName = _argument => {
    const reg = /^\$\.+/;
    const item = _.find(_argument, o => reg.test(o.value));
    // NOTE: get integration name from $.${namespace}.${integration}/data.integration/sonarQube.server
    if (item) {
      const value = _.get(item, 'value').split('/data.integration');
      const integration = value[0].split('.')[2];
      return integration;
    }
  };

  editResource = (r, index) => {
    let state = {
      modifyData: r,
      visible: true,
      modifyIndex: index,
    };
    this.setState(state);
  };
  render() {
    const {
      resources,
      resourcesField,
      type = 'inputs',
      update,
      projectName,
    } = this.props;
    const colSpan = type === 'inputs' ? 6 : 9;
    const { visible, modifyData, modifyIndex } = this.state;
    return (
      <FormItem label={intl.get('sideNav.resource')} {...drawerFormItemLayout}>
        <FieldArray
          name={resourcesField}
          render={arrayHelpers => (
            <div>
              {resources.length > 0 && (
                <Row gutter={16}>
                  <Col span={colSpan}>{intl.get('name')}</Col>
                  <Col span={colSpan}>{intl.get('type')}</Col>
                  {type === 'inputs' && (
                    <Col span={colSpan}>{intl.get('path')}</Col>
                  )}
                </Row>
              )}
              {/* TODO(qme): click resource list show modal and restore resource form */}
              {resources.map((r, i) => {
                return (
                  <FastField
                    key={i}
                    name={`${resourcesField}.${i}`}
                    validate={value =>
                      !_.get(value, 'name')
                        ? { name: intl.get('validate.required') }
                        : undefined
                    }
                    render={({ field, form }) => {
                      const error = _.get(
                        form,
                        `errors.${resourcesField}.${i}.name`
                      );
                      const itemStyle = error
                        ? { border: '1px solid #f5222d', borderRadius: '4px' }
                        : {};
                      return (
                        <Fragment>
                          <div style={itemStyle}>
                            <Row gutter={16}>
                              <Col span={colSpan}>
                                {_.get(r, 'name') || '--'}
                              </Col>
                              <Col span={colSpan}>{_.get(r, 'type')}</Col>
                              {type === 'inputs' && (
                                <Col span={colSpan}>{_.get(r, 'path')}</Col>
                              )}
                              <Col span={6}>
                                <div className={style['resource-action']}>
                                  <Button
                                    type="circle"
                                    icon="edit"
                                    onClick={() => {
                                      this.editResource(r, i);
                                    }}
                                  />
                                  <Button
                                    type="circle"
                                    icon="delete"
                                    onClick={() => arrayHelpers.remove(i)}
                                  />
                                </div>
                              </Col>
                            </Row>
                          </div>
                          {error && (
                            <div style={{ color: '#f5222d' }}>
                              {intl.get('validate.incompleteResource')}
                            </div>
                          )}
                        </Fragment>
                      );
                    }}
                  />
                );
              })}
              <Button
                ico="plus"
                onClick={() => {
                  this.setState({ visible: true });
                }}
              >
                {intl.get('workflow.addResource')}
              </Button>
              {visible && (
                <Resource
                  type={type}
                  handleModalClose={() => {
                    this.setState({ visible: false, modifyData: null });
                  }}
                  SetReasourceValue={(value, modify) => {
                    modify
                      ? arrayHelpers.replace(modifyIndex, value)
                      : arrayHelpers.push(value);
                  }}
                  visible={visible}
                  update={update}
                  projectName={projectName}
                  modifyData={modifyData}
                />
              )}
            </div>
          )}
        />
      </FormItem>
    );
  }
}

export default ResourceArray;
