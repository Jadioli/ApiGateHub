import React from 'react';
import { Alert } from 'antd';

export default class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false, error: null, errorInfo: null };
  }

  static getDerivedStateFromError(error) {
    return { hasError: true };
  }

  componentDidCatch(error, errorInfo) {
    this.setState({ error, errorInfo });
    console.error("ErrorBoundary caught an error:", error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div style={{ padding: 24, textAlign: 'left' }}>
          <Alert
            type="error"
            message="Something went wrong rendering this component."
            description={
              <pre style={{ whiteSpace: 'pre-wrap', fontSize: 12 }}>
                {this.state.error?.toString()}
                <br />
                {this.state.errorInfo?.componentStack}
              </pre>
            }
          />
        </div>
      );
    }
    return this.props.children;
  }
}
