/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha2

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	infrav1alpha3 "sigs.k8s.io/cluster-api-provider-aws/api/v1alpha3"
)

func TestConvertAWSMachine(t *testing.T) {
	g := NewWithT(t)

	t.Run("from hub", func(t *testing.T) {
		t.Run("should restore SecretARN, assuming old version of object without field", func(t *testing.T) {
			src := &infrav1alpha3.AWSMachine{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: infrav1alpha3.AWSMachineSpec{
					CloudInit: infrav1alpha3.CloudInit{
						InsecureSkipSecretsManager: true,
						SecretARN:                  "something",
					},
				},
			}
			dst := &AWSMachine{}
			g.Expect(dst.ConvertFrom(src)).To(Succeed())
			restored := &infrav1alpha3.AWSMachine{}
			g.Expect(dst.ConvertTo(restored)).To(Succeed())
			g.Expect(restored.Spec.CloudInit.SecretARN).To(Equal(src.Spec.CloudInit.SecretARN))
			g.Expect(restored.Spec.CloudInit.InsecureSkipSecretsManager).To(Equal(src.Spec.CloudInit.InsecureSkipSecretsManager))
		})
	})
	t.Run("should prefer newer cloudinit data on the v1alpha2 obj", func(t *testing.T) {
		src := &infrav1alpha3.AWSMachine{
			ObjectMeta: metav1.ObjectMeta{},
			Spec: infrav1alpha3.AWSMachineSpec{
				CloudInit: infrav1alpha3.CloudInit{
					SecretARN: "something",
				},
			},
		}
		dst := &AWSMachine{
			Spec: AWSMachineSpec{
				CloudInit: &CloudInit{
					EnableSecureSecretsManager: true,
					SecretARN:                  "something-else",
				},
			},
		}
		g.Expect(dst.ConvertFrom(src)).To(Succeed())
		restored := &infrav1alpha3.AWSMachine{}
		g.Expect(dst.ConvertTo(restored)).To(Succeed())
		g.Expect(restored.Spec.CloudInit.SecretARN).To(Equal(src.Spec.CloudInit.SecretARN))
	})
	t.Run("should restore ImageLookupBaseOS", func(t *testing.T) {
		src := &infrav1alpha3.AWSMachine{
			ObjectMeta: metav1.ObjectMeta{},
			Spec: infrav1alpha3.AWSMachineSpec{
				ImageLookupBaseOS: "amazon-linux",
			},
		}
		dst := &AWSMachine{}
		g.Expect(dst.ConvertFrom(src)).To(Succeed())
		restored := &infrav1alpha3.AWSMachine{}
		g.Expect(dst.ConvertTo(restored)).To(Succeed())
		g.Expect(restored.Spec.ImageLookupBaseOS).To(Equal(src.Spec.ImageLookupBaseOS))
	})
}
