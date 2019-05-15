import { Button, Row, Col, Form } from 'antd';
import { FieldArray } from 'formik';
import { defaultFormItemLayout } from '@/lib/const';
import Resource from './Form';
import { inject, observer } from 'mobx-react';
import PropTypes from 'prop-types';

const FormItem = Form.Item;

@inject('resource')
@observer
class ResourceArray extends React.Component {
  static propTypes = {
    resources: PropTypes.array,
    resourcesField: PropTypes.string,
    type: PropTypes.oneOf(['inputs', 'outputs']),
  };

  state = {
    visible: false,
  };

  editResource = r => {
    const {
      update,
      project,
      resource: { getResource },
    } = this.props;
    let state = {
      modifyData: r,
      visible: true,
    };
    console.log('rrrr', r, update);

    if (update) {
      getResource(project, r.name, data => {
        console.log('^^^^^');
        const info = _.merge(
          _.pick(data, ['metadata.name', 'spec']),
          _.pick(r, ['type', 'path'])
        );
        state.modifyData = info;
      });
    }
    console.log('qmeqme', state.modifyData);
    this.setState(state);
  };
  render() {
    const {
      resources,
      resourcesField,
      type = 'inputs',
      update,
      project,
    } = this.props;
    const { visible, modifyData } = this.state;
    console.log('modifyData', modifyData);
    return (
      <FormItem label={intl.get('sideNav.resource')} {...defaultFormItemLayout}>
        <FieldArray
          name={resourcesField}
          render={arrayHelpers => (
            <div>
              {resources.length > 0 && (
                <Row gutter={16}>
                  <Col span={10}>{intl.get('name')}</Col>
                  <Col span={10}>{intl.get('path')}</Col>
                </Row>
              )}
              {/* TODO(qme): click resource list show modal and restore resource form */}
              {resources.map((r, i) => (
                <Row gutter={16} key={i}>
                  <Col span={10}>{r.name}</Col>
                  <Col span={10}>{r.path}</Col>
                  <Col span={4}>
                    <Button
                      type="circle"
                      icon="edit"
                      onClick={() => {
                        this.editResource(r);
                        // console.log('r', r);
                        // this.setState({ visible: true, modifyData: r });
                      }}
                    />
                    <Button
                      type="circle"
                      icon="delete"
                      onClick={() => arrayHelpers.remove(i)}
                    />
                  </Col>
                </Row>
              ))}
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
                    this.setState({ visible: false });
                  }}
                  SetReasourceValue={value => {
                    arrayHelpers.push(value);
                  }}
                  visible={visible}
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
