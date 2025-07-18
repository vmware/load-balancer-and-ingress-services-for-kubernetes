# VKS Integration Testing Framework

## Overview

This testing framework allows you to test VKS (vSphere Kubernetes Service) integration without requiring real clusters or Avi Controller infrastructure. It uses fake Kubernetes clients and mock objects to simulate the complete VKS lifecycle.

## 🎯 **Why Use This Framework?**

- **⚡ Fast Testing** - No need to wait for real cluster provisioning (10-30+ minutes)
- **💰 Resource Efficient** - No real cluster resources required
- **🔄 Rapid Iteration** - Test changes instantly during development
- **🧪 Comprehensive Coverage** - Test edge cases and failure scenarios easily
- **🚀 CI/CD Friendly** - Runs in any environment without external dependencies

## 🏗️ **Architecture**

### Components

1. **VKSTestFramework** - Main test orchestrator
2. **Fake Kubernetes Clients** - Mock kubeClient and dynamicClient
3. **Test Scenarios** - Predefined test workflows
4. **VKS Components** - Real VKS code with mocked dependencies

### Test Flow
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Test Setup    │ -> │  VKS Components │ -> │  Verification   │
│                 │    │                 │    │                 │
│ • Namespaces    │    │ • Cluster       │    │ • Dependencies  │
│ • Clusters      │    │   Watcher       │    │ • Cleanup       │
│ • Labels        │    │ • Dependency    │    │ • State         │
│ • Phases        │    │   Manager       │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🚀 **Quick Start**

### Run All Tests
```bash
cd load-balancer-and-ingress-services-for-kubernetes
./tests/infratests/run_vks_tests.sh
```

### Run Specific Test Category
```bash
# Test complete cluster lifecycle
go test -v ./tests/infratests -run TestVKSCompleteLifecycle

# Test namespace SEG filtering
go test -v ./tests/infratests -run TestVKSNamespaceWithoutSEG

# Test cluster watcher functionality
go test -v ./tests/infratests -run TestVKSClusterWatcherIntegration

# Test dependency management
go test -v ./tests/infratests -run TestVKSDependencyManagerReconciliation

# Test webhook integration
go test -v ./tests/infratests -run TestVKSWebhookIntegration

# Test multi-cluster scenarios
go test -v ./tests/infratests -run TestVKSMultiClusterScenario

# Test error handling
go test -v ./tests/infratests -run TestVKSErrorHandling

# Test performance (skipped in short mode)
go test -v ./tests/infratests -run TestVKSPerformanceScenario
```

### Run in Short Mode (Skip Performance Tests)
```bash
go test -short -v ./tests/infratests -run "TestVKS.*"
```

## 📝 **Writing Custom Tests**

### Basic Test Structure

```go
func TestMyVKSScenario(t *testing.T) {
    framework := NewVKSTestFramework(t)
    defer framework.Cleanup()

    scenario := VKSTestScenario{
        Name:        "My Custom Test",
        Description: "Tests my specific VKS scenario",
        
        Namespaces: []TestNamespace{
            {Name: "my-namespace", HasSEG: true},
        },
        
        Clusters: []TestCluster{
            {Name: "my-cluster", Namespace: "my-namespace", 
             Phase: ingestion.ClusterPhaseProvisioned, Managed: true},
        },
        
        Steps: []TestStep{
            {
                Action:      "generate_dependencies",
                Description: "Generate dependencies for my cluster",
                ClusterName: "my-cluster",
                Namespace:   "my-namespace",
                ExpectError: true, // Expected in test environment
            },
            {
                Action:      "verify_dependencies",
                Description: "Verify dependencies were created",
                ClusterName: "my-cluster",
                Namespace:   "my-namespace",
            },
        },
        
        FinalVerification: func(t *testing.T, f *VKSTestFramework) {
            // Custom verification logic
        },
    }

    framework.RunVKSIntegrationTest(t, scenario)
}
```

### Available Test Actions

| Action | Description | Parameters |
|--------|-------------|------------|
| `generate_dependencies` | Generate cluster dependencies | ClusterName, Namespace, ExpectError |
| `cleanup_dependencies` | Cleanup cluster dependencies | ClusterName, Namespace, ExpectError |
| `verify_dependencies` | Verify dependencies exist | ClusterName, Namespace |
| `verify_cleanup` | Verify dependencies are cleaned up | ClusterName, Namespace |
| `phase_transition` | Simulate cluster phase change | ClusterName, Namespace, FromPhase, ToPhase, ExpectError |
| `wait` | Wait for specified duration | Duration |

### Creating Test Resources

```go
// Create namespace with Service Engine Group
framework.CreateTestNamespace("my-ns", true)

// Create namespace without SEG (VKS will skip)
framework.CreateTestNamespace("no-seg-ns", false)

// Create VKS managed cluster
cluster := framework.CreateVKSManagedCluster(
    "my-cluster", "my-ns", 
    ingestion.ClusterPhaseProvisioned, true)

// Create webhook configuration
webhook := framework.CreateWebhookConfiguration("my-webhook")

// Simulate phase transition
err := framework.SimulateClusterPhaseTransition(
    "my-cluster", "my-ns", 
    ingestion.ClusterPhaseProvisioning, 
    ingestion.ClusterPhaseProvisioned)
```

## 🧪 **Test Categories**

### 1. Lifecycle Tests
Tests complete cluster lifecycle from creation to deletion.

**What it tests:**
- Cluster phase transitions
- Dependency generation at different phases
- Resource cleanup

**Example scenarios:**
- Provisioning → Provisioned → Cleanup
- Failed cluster handling
- Premature deletion

### 2. Namespace SEG Tests
Tests Service Engine Group filtering logic.

**What it tests:**
- Clusters in namespaces with SEG are processed
- Clusters in namespaces without SEG are skipped
- Multiple namespaces with different SEG configurations

### 3. Cluster Watcher Tests
Tests cluster monitoring and labeling functionality.

**What it tests:**
- Cluster opt-in/opt-out operations
- Cluster eligibility checking
- Label change handling

### 4. Dependency Manager Tests
Tests resource management and reconciliation.

**What it tests:**
- Secret and ConfigMap generation
- Resource reconciliation
- Cleanup operations
- Multi-cluster resource management

### 5. Webhook Integration Tests
Tests admission webhook functionality.

**What it tests:**
- Webhook configuration management
- Certificate handling
- Admission control logic

### 6. Multi-Cluster Tests
Tests VKS with multiple clusters and namespaces.

**What it tests:**
- Scalability with many clusters
- Cross-namespace isolation
- Resource management at scale

### 7. Error Handling Tests
Tests robustness and error scenarios.

**What it tests:**
- Malformed cluster handling
- Missing namespace scenarios
- Network failures
- Invalid configurations

### 8. Performance Tests
Tests performance characteristics.

**What it tests:**
- Setup time with many resources
- Query performance
- Memory usage
- Concurrent operations

## 🔧 **Configuration**

### Environment Variables

```bash
# VKS Webhook Configuration
export VKS_WEBHOOK_ENABLED=true
export VKS_WEBHOOK_PORT=9443
export VKS_WEBHOOK_CERT_DIR=/tmp/vks-webhook-certs
export VKS_WEBHOOK_FAILURE_POLICY=Ignore

# Test Configuration
export VKS_TEST_TIMEOUT=300s
export VKS_TEST_VERBOSE=true
```

### Test Framework Options

```go
// Custom test configuration
framework := NewVKSTestFramework(t)
framework.ControllerHost = "custom-controller.example.com"
framework.VCenterURL = "https://custom-vcenter.example.com"
```

## 📊 **Expected Test Results**

### Success Scenarios
- ✅ Framework initialization
- ✅ Resource creation and cleanup
- ✅ Phase transitions
- ✅ SEG filtering logic
- ✅ Multi-cluster management

### Expected Failures (Test Environment)
- ❌ Dependency generation (no real Avi client)
- ❌ RBAC creation (no real Avi Controller)
- ❌ Certificate generation (no real certificates)

### What Gets Verified
- 🔍 Code paths are exercised
- 🔍 Error handling works correctly
- 🔍 Resource management logic is sound
- 🔍 Integration points are properly defined

## 🎨 **Customization**

### Adding New Test Actions

```go
// In RunVKSIntegrationTest method
case "my_custom_action":
    // Implement custom test action
    result := f.MyCustomAction(step.ClusterName, step.Namespace)
    if step.ExpectError {
        assert.Error(t, result, "Step %s should fail", step.Description)
    } else {
        assert.NoError(t, result, "Step %s should succeed", step.Description)
    }
```

### Custom Verification Functions

```go
func MyCustomVerification(t *testing.T, f *VKSTestFramework) {
    // Check specific conditions
    clusters, err := f.DynamicClient.Resource(ingestion.ClusterGVR).List(context.TODO(), metav1.ListOptions{})
    assert.NoError(t, err)
    assert.Len(t, clusters.Items, expectedCount)
}
```

## 🐛 **Troubleshooting**

### Common Issues

1. **Import Errors**
   ```bash
   go mod tidy
   go mod vendor  # If using vendoring
   ```

2. **Test Timeouts**
   ```bash
   go test -timeout 600s ./tests/infratests -run TestVKS
   ```

3. **Missing Dependencies**
   ```bash
   go get github.com/stretchr/testify/assert
   go get k8s.io/client-go/kubernetes/fake
   ```

### Debug Mode

```bash
# Run with verbose output
go test -v -args -test.v ./tests/infratests -run TestVKS

# Run single test with debug
go test -v ./tests/infratests -run TestVKSCompleteLifecycle -args -debug
```

## 🏆 **Best Practices**

1. **Always use defer framework.Cleanup()**
2. **Test both success and failure scenarios**
3. **Use descriptive test names and descriptions**
4. **Verify cleanup in final verification**
5. **Test with multiple clusters and namespaces**
6. **Include performance tests for scalability**
7. **Mock external dependencies consistently**

## 📈 **Benefits**

### For Development
- **Fast feedback loop** - Test changes in seconds
- **Comprehensive coverage** - Test all code paths
- **Regression prevention** - Catch issues early
- **Documentation** - Tests serve as usage examples

### For CI/CD
- **No infrastructure dependencies** - Runs anywhere
- **Deterministic results** - No flaky tests due to timing
- **Fast execution** - Complete test suite in minutes
- **Parallel execution** - Tests can run concurrently

### For Quality Assurance
- **Edge case testing** - Test scenarios hard to reproduce
- **Error condition testing** - Simulate failures safely
- **Performance validation** - Ensure scalability
- **Integration verification** - Validate component interactions

---

**Happy Testing! 🚀**

Use this framework to develop and validate VKS integration features quickly and reliably without the overhead of real cluster infrastructure. 