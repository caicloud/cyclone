import React from 'react';
import { RedocStandalone } from 'redoc';

class Swagger extends React.Component {
  render() {
    return (
      <RedocStandalone
        specUrl="/api.v1alpha1.json"
        options={{
          nativeScrollbars: true,
          theme: { colors: { main: '#dd5522' } },
        }}
      />
    );
  }
}

export default Swagger;
