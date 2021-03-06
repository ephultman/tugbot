package container

import (
	"errors"
	"strings"
	"testing"

	"github.com/samalba/dockerclient"
	"github.com/samalba/dockerclient/mockclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func allContainers(Container) bool {
	return true
}

func noContainers(Container) bool {
	return false
}

func TestListContainers_Success(t *testing.T) {
	ci := &dockerclient.ContainerInfo{Image: "abc123", Config: &dockerclient.ContainerConfig{Image: "img"}}
	ii := &dockerclient.ImageInfo{}
	api := mockclient.NewMockClient()
	api.On("ListContainers", true, false, "").Return([]dockerclient.Container{{Id: "foo", Names: []string{"bar"}}}, nil)
	api.On("InspectContainer", "foo").Return(ci, nil)
	api.On("InspectImage", "abc123").Return(ii, nil)

	client := dockerClient{api: api}
	cs, err := client.ListContainers(allContainers)

	assert.NoError(t, err)
	assert.Len(t, cs, 1)
	assert.Equal(t, ci, cs[0].containerInfo)
	assert.Equal(t, ii, cs[0].imageInfo)
	api.AssertExpectations(t)
}

func TestListContainers_Filter(t *testing.T) {
	ci := &dockerclient.ContainerInfo{Image: "abc123", Config: &dockerclient.ContainerConfig{Image: "img"}}
	ii := &dockerclient.ImageInfo{}
	api := mockclient.NewMockClient()
	api.On("ListContainers", true, false, "").Return([]dockerclient.Container{{Id: "foo", Names: []string{"bar"}}}, nil)
	api.On("InspectContainer", "foo").Return(ci, nil)
	api.On("InspectImage", "abc123").Return(ii, nil)

	client := dockerClient{api: api}
	cs, err := client.ListContainers(noContainers)

	assert.NoError(t, err)
	assert.Len(t, cs, 0)
	api.AssertExpectations(t)
}

func TestListContainers_ListError(t *testing.T) {
	api := mockclient.NewMockClient()
	api.On("ListContainers", true, false, "").Return([]dockerclient.Container{}, errors.New("oops"))

	client := dockerClient{api: api}
	_, err := client.ListContainers(allContainers)

	assert.Error(t, err)
	assert.EqualError(t, err, "oops")
	api.AssertExpectations(t)
}

func TestListContainers_InspectContainerError(t *testing.T) {
	api := mockclient.NewMockClient()
	api.On("ListContainers", true, false, "").Return([]dockerclient.Container{{Id: "foo", Names: []string{"bar"}}}, nil)
	api.On("InspectContainer", "foo").Return(&dockerclient.ContainerInfo{}, errors.New("uh-oh"))

	client := dockerClient{api: api}
	cs, err := client.ListContainers(allContainers)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(cs))
	api.AssertExpectations(t)
}

func TestListContainers_InspectImageError(t *testing.T) {
	ci := &dockerclient.ContainerInfo{Image: "abc123", Config: &dockerclient.ContainerConfig{Image: "img"}}
	ii := &dockerclient.ImageInfo{}
	api := mockclient.NewMockClient()
	api.On("ListContainers", true, false, "").Return([]dockerclient.Container{{Id: "foo", Names: []string{"bar"}}}, nil)
	api.On("InspectContainer", "foo").Return(ci, nil)
	api.On("InspectImage", "abc123").Return(ii, errors.New("whoops"))

	client := dockerClient{api: api}
	cs, err := client.ListContainers(allContainers)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(cs))
	api.AssertExpectations(t)
}

func TestStartContainerFrom_Success(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			Name:       "foo",
			Config:     &dockerclient.ContainerConfig{},
			HostConfig: &dockerclient.HostConfig{},
		},
		imageInfo: &dockerclient.ImageInfo{
			Config: &dockerclient.ContainerConfig{},
		},
	}

	api := mockclient.NewMockClient()
	api.On("CreateContainer",
		mock.MatchedBy(func(config *dockerclient.ContainerConfig) bool {
			return config.Labels[TugbotCreatedFrom] == "foo"
		}),
		mock.MatchedBy(func(name string) bool {
			return strings.HasPrefix(name, "tugbot_foo_")
		}),
		mock.AnythingOfType("*dockerclient.AuthConfig")).Return("def789", nil).Once()
	api.On("StartContainer", "def789", mock.AnythingOfType("*dockerclient.HostConfig")).Return(nil).Once()

	client := dockerClient{api: api}
	err := client.StartContainerFrom(c)

	assert.NoError(t, err)
	api.AssertExpectations(t)
}

func TestStartContainerFrom_CreateContainerError(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			Name:       "foo",
			Config:     &dockerclient.ContainerConfig{},
			HostConfig: &dockerclient.HostConfig{},
		},
		imageInfo: &dockerclient.ImageInfo{
			Config: &dockerclient.ContainerConfig{},
		},
	}

	api := mockclient.NewMockClient()
	api.On("CreateContainer",
		mock.MatchedBy(func(config *dockerclient.ContainerConfig) bool {
			return config.Labels[TugbotCreatedFrom] == "foo"
		}),
		mock.MatchedBy(func(name string) bool {
			return strings.HasPrefix(name, "tugbot_foo_")
		}), mock.AnythingOfType("*dockerclient.AuthConfig")).Return("", errors.New("oops")).Once()

	client := dockerClient{api: api}
	err := client.StartContainerFrom(c)

	assert.Error(t, err)
	assert.EqualError(t, err, "oops")
	api.AssertExpectations(t)
}

func TestStartContainerFrom_StartContainerError(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			Name:       "foo",
			Config:     &dockerclient.ContainerConfig{},
			HostConfig: &dockerclient.HostConfig{},
		},
		imageInfo: &dockerclient.ImageInfo{
			Config: &dockerclient.ContainerConfig{},
		},
	}

	api := mockclient.NewMockClient()
	api.On("CreateContainer",
		mock.MatchedBy(func(config *dockerclient.ContainerConfig) bool {
			return config.Labels[TugbotCreatedFrom] == "foo"
		}),
		mock.MatchedBy(func(name string) bool {
			return strings.HasPrefix(name, "tugbot_foo_")
		}),
		mock.AnythingOfType("*dockerclient.AuthConfig")).Return("created-container-id", nil).Once()
	api.On("StartContainer", "created-container-id", mock.Anything).Return(errors.New("whoops")).Once()

	client := dockerClient{api: api}
	err := client.StartContainerFrom(c)

	assert.Error(t, err)
	assert.EqualError(t, err, "whoops")
	api.AssertExpectations(t)
}
