package mkcontainer_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/SAP/xp-clifford/mkcontainer"
)



type testOWG struct {
	guid string
}

func (towg *testOWG) GetGUID() string {
	return towg.guid
}

var _ mkcontainer.ItemWithGUID = &testOWG{}

type testOWN struct {
	name string
}

func (town *testOWN) GetName() string {
	return town.name
}

var _ mkcontainer.ItemWithName = &testOWN{}

var _ = Describe("A Container", func() {
	var cnt mkcontainer.Container
	BeforeEach(func(){
		cnt = mkcontainer.New()
	})
	Describe("after initialization", func(){
		It("does not contain object with GUID 'guid1'", func(){
			Expect(cnt.GetByGUID("guid1")).To(BeNil())
		})
		It("returns an empty list of GUIDs", func(){
			Expect(cnt.GetGUIDs()).To(BeEmpty())
		})
		It("can be iterated by GUID", func(){
			Expect(cnt.AllByGUIDs()).To(BeEmpty())
		})
		It("does not contain object with name 'name1'", func(){
			Expect(cnt.GetByName("name1")).To(BeNil())
		})
		It("returns an empyt list of names", func(){
			Expect(cnt.GetNames()).To(BeEmpty())
		})
		It("returns an empty container", func(){
			Expect(cnt.IsEmpty()).To(BeTrue())
		})
	})
	Describe("after storing a single ObjectWithGUID", func(){
		var owg mkcontainer.Item

		BeforeEach(func(){
			owg = &testOWG{"guid1"}
			cnt.Store(owg)
		})
		It("can return it by GUID", func(){
			Expect(cnt.GetByGUID("guid1")).To(Equal(owg))
			Expect(cnt.GetByGUID("unknown")).To(BeNil())
		})
		It("can list the single GUID", func(){
			Expect(cnt.GetGUIDs()).To(Equal([]string{"guid1"}))
		})
		It("can be iterated by GUID", func(){
			Expect(cnt.AllByGUIDs()).To(HaveLen(1))
			Expect(cnt.AllByGUIDs()).To(HaveKeyWithValue("guid1", owg))
		})
		It("does not contain object with name 'name1'", func(){
			Expect(cnt.GetByName("name1")).To(BeNil())
		})
		It("returns an empty list of names", func(){
			Expect(cnt.GetNames()).To(BeEmpty())
		})
		It("returns a non-empty container", func(){
			Expect(cnt.IsEmpty()).To(BeFalse())
		})
	})
	Describe("after storing a multiple ObjectWithGUIDs", func(){
		var owg []mkcontainer.Item

		BeforeEach(func(){
			owg = []mkcontainer.Item{
				&testOWG{"guid1"},
				&testOWG{"guid2"},
				&testOWG{"guid3"},
			}
			cnt.Store(owg...)
		})
		It("can return the elements by GUID", func(){
			Expect(cnt.GetByGUID("guid1")).To(Equal(owg[0]))
			Expect(cnt.GetByGUID("guid2")).To(Equal(owg[1]))
			Expect(cnt.GetByGUID("guid3")).To(Equal(owg[2]))
			Expect(cnt.GetByGUID("unknown")).To(BeNil())
		})
		It("can list the GUIDs", func(){
			Expect(cnt.GetGUIDs()).To(Equal([]string{
				"guid1",
				"guid2",
				"guid3",
			}))
		})
		It("can be iterated by GUID", func(){
			Expect(cnt.AllByGUIDs()).To(HaveLen(3))
			Expect(cnt.AllByGUIDs()).To(HaveKeyWithValue("guid1", owg[0]))
			Expect(cnt.AllByGUIDs()).To(HaveKeyWithValue("guid2", owg[1]))
			Expect(cnt.AllByGUIDs()).To(HaveKeyWithValue("guid3", owg[2]))
		})
		It("does not contain object with name 'name1'", func(){
			Expect(cnt.GetByName("name1")).To(BeNil())
		})
		It("returns an empyt list of names", func(){
			Expect(cnt.GetNames()).To(BeEmpty())
		})
		It("returns a non-empty container", func(){
			Expect(cnt.IsEmpty()).To(BeFalse())
		})
	})
	Describe("after storing a single ObjectWithName", func(){
		var own mkcontainer.Item

		BeforeEach(func(){
			own = &testOWN{"name1"}
			cnt.Store(own)
		})
		It("does not contain object with GUID 'guid1'", func(){
			Expect(cnt.GetByGUID("guid1")).To(BeNil())
		})
		It("returns an empty list of GUIDs", func(){
			Expect(cnt.GetGUIDs()).To(BeEmpty())
		})
		It("can be iterated by GUID", func(){
			Expect(cnt.AllByGUIDs()).To(BeEmpty())
		})
		It("can return it by name", func(){
			Expect(cnt.GetByName("name1")).To(Equal([]mkcontainer.ItemWithName{
				own.(mkcontainer.ItemWithName),
			}))
			Expect(cnt.GetByName("unknown")).To(BeNil())
		})
		It("can list the single name", func(){
			Expect(cnt.GetNames()).To(Equal([]string{"name1"}))
		})
		It("can be iterated by names", func(){
			Expect(cnt.AllByNames()).To(HaveLen(1))

			Expect(cnt.AllByNames()).To(HaveKeyWithValue("name1", []mkcontainer.ItemWithName{
				own.(mkcontainer.ItemWithName),
			}))
		})
		It("returns a non-empty container", func(){
			Expect(cnt.IsEmpty()).To(BeFalse())
		})
	})
	Describe("after storing multiple ObjectWithName", func(){
		var own []mkcontainer.Item

		BeforeEach(func(){
			own = []mkcontainer.Item{
				&testOWN{"name1"},
				&testOWN{"name2"},
				&testOWN{"name1"},
			}
			cnt.Store(own...)
		})
		It("does not contain object with GUID 'guid1'", func(){
			Expect(cnt.GetByGUID("guid1")).To(BeNil())
		})
		It("returns an empty list of GUIDs", func(){
			Expect(cnt.GetGUIDs()).To(BeEmpty())
		})
		It("can be iterated by GUID", func(){
			Expect(cnt.AllByGUIDs()).To(BeEmpty())
		})
		It("can return it by name", func(){
			Expect(cnt.GetByName("name1")).To(Equal([]mkcontainer.ItemWithName{
				own[0].(mkcontainer.ItemWithName),
				own[2].(mkcontainer.ItemWithName),
			}))
			Expect(cnt.GetByName("name2")).To(Equal([]mkcontainer.ItemWithName{
				own[1].(mkcontainer.ItemWithName),
			}))
			Expect(cnt.GetByName("unknown")).To(BeNil())
		})
		It("can list the names", func(){
			Expect(cnt.GetNames()).To(Equal([]string{
				"name1", "name2",
			}))
		})
		It("can be iterated by names", func(){
			Expect(cnt.AllByNames()).To(HaveLen(2))
			Expect(cnt.AllByNames()).To(HaveKeyWithValue("name1", []mkcontainer.ItemWithName{
				own[0].(mkcontainer.ItemWithName),
				own[2].(mkcontainer.ItemWithName),
			}))
			Expect(cnt.AllByNames()).To(HaveKeyWithValue("name2", []mkcontainer.ItemWithName{
				own[1].(mkcontainer.ItemWithName),
			}))
		})
		It("returns a non-empty container", func(){
			Expect(cnt.IsEmpty()).To(BeFalse())
		})
	})
	Describe("after storing multiple mixed Objects", func(){
		var own []mkcontainer.Item

		BeforeEach(func(){
			own = []mkcontainer.Item{
				&testOWN{"name1"},
				&testOWN{"name2"},
				&testOWN{"name1"},
				&testOWG{"guid1"},
				&testOWG{"guid2"},
				&testOWG{"guid3"},
			}
			cnt.Store(own...)
		})

		It("can return the elements by GUID", func(){
			Expect(cnt.GetByGUID("guid1")).To(Equal(own[3]))
			Expect(cnt.GetByGUID("guid2")).To(Equal(own[4]))
			Expect(cnt.GetByGUID("guid3")).To(Equal(own[5]))
			Expect(cnt.GetByGUID("unknown")).To(BeNil())
		})
		It("can list the GUIDs", func(){
			Expect(cnt.GetGUIDs()).To(Equal([]string{
				"guid1",
				"guid2",
				"guid3",
			}))
		})
		It("can be iterated by GUID", func(){
			Expect(cnt.AllByGUIDs()).To(HaveLen(3))
			Expect(cnt.AllByGUIDs()).To(HaveKeyWithValue("guid1", own[3]))
			Expect(cnt.AllByGUIDs()).To(HaveKeyWithValue("guid2", own[4]))
			Expect(cnt.AllByGUIDs()).To(HaveKeyWithValue("guid3", own[5]))
		})
		It("can return it by name", func(){
			Expect(cnt.GetByName("name1")).To(Equal([]mkcontainer.ItemWithName{
				own[0].(mkcontainer.ItemWithName),
				own[2].(mkcontainer.ItemWithName),
			}))
			Expect(cnt.GetByName("name2")).To(Equal([]mkcontainer.ItemWithName{
				own[1].(mkcontainer.ItemWithName),
			}))
			Expect(cnt.GetByName("unknown")).To(BeNil())
		})
		It("can list the names", func(){
			Expect(cnt.GetNames()).To(Equal([]string{
				"name1", "name2",
			}))
		})
		It("can be iterated by names", func(){
			Expect(cnt.AllByNames()).To(HaveLen(2))
			Expect(cnt.AllByNames()).To(HaveKeyWithValue("name1", []mkcontainer.ItemWithName{
				own[0].(mkcontainer.ItemWithName),
				own[2].(mkcontainer.ItemWithName),
			}))
			Expect(cnt.AllByNames()).To(HaveKeyWithValue("name2", []mkcontainer.ItemWithName{
				own[1].(mkcontainer.ItemWithName),
			}))
		})
		It("returns a non-empty container", func(){
			Expect(cnt.IsEmpty()).To(BeFalse())
		})
	})
})
