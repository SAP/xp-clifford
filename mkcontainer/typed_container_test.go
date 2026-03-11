package mkcontainer_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/SAP/xp-clifford/mkcontainer"
)

var _ = Describe("A TypedContainer", func() {
	var cnt mkcontainer.TypedContainer[mkcontainer.Item]
	var cntg mkcontainer.TypedContainer[*testOWG]
	var cntn mkcontainer.TypedContainer[*testOWN]
	BeforeEach(func() {
		cnt = mkcontainer.NewTyped[mkcontainer.Item]()
		cntg = mkcontainer.NewTyped[*testOWG]()
		cntn = mkcontainer.NewTyped[*testOWN]()
	})
	Describe("after initialization", func() {
		It("does not contain object with GUID 'guid1'", func() {
			Expect(cntg.GetByGUID("guid1")).To(BeNil())
		})
		It("returns an empty list of GUIDs", func() {
			Expect(cntg.GetGUIDs()).To(BeEmpty())
		})
		It("can be iterated by GUID", func() {
			Expect(cntg.AllByGUIDs()).To(BeEmpty())
		})
		It("does not contain object with name 'name1'", func() {
			Expect(cntg.GetByName("name1")).To(BeNil())
		})
		It("returns an empyt list of names", func() {
			Expect(cntg.GetNames()).To(BeEmpty())
		})
		It("is empty", func() {
			Expect(cntg.IsEmpty()).To(BeTrue())
		})
	})
	Describe("after storing a single ObjectWithGUID", func() {
		var owg *testOWG

		BeforeEach(func() {
			owg = &testOWG{"guid1"}
			cntg.Store(owg)
		})
		It("can return it by GUID", func() {
			Expect(cntg.GetByGUID("guid1")).To(Equal(owg))
			Expect(cntg.GetByGUID("unknown")).To(BeNil())
		})
		It("can list the single GUID", func() {
			Expect(cntg.GetGUIDs()).To(Equal([]string{"guid1"}))
		})
		It("can be iterated by GUID", func() {
			Expect(cntg.AllByGUIDs()).To(HaveLen(1))
			Expect(cntg.AllByGUIDs()).To(HaveKeyWithValue("guid1", owg))
		})
		It("does not contain object with name 'name1'", func() {
			Expect(cntg.GetByName("name1")).To(BeNil())
		})
		It("returns an empyt list of names", func() {
			Expect(cntg.GetNames()).To(BeEmpty())
		})
		It("is not empty", func() {
			Expect(cntg.IsEmpty()).To(BeFalse())
		})
	})
	Describe("after storing a multiple ObjectWithGUIDs", func() {
		var owg []*testOWG

		BeforeEach(func() {
			owg = []*testOWG{
				&testOWG{"guid1"},
				&testOWG{"guid2"},
				&testOWG{"guid3"},
			}
			cntg.Store(owg...)
		})
		It("can return the elements by GUID", func() {
			Expect(cntg.GetByGUID("guid1")).To(Equal(owg[0]))
			Expect(cntg.GetByGUID("guid2")).To(Equal(owg[1]))
			Expect(cntg.GetByGUID("guid3")).To(Equal(owg[2]))
			Expect(cntg.GetByGUID("unknown")).To(BeNil())
		})
		It("can list the GUIDs", func() {
			Expect(cntg.GetGUIDs()).To(Equal([]string{
				"guid1",
				"guid2",
				"guid3",
			}))
		})
		It("can be iterated by GUID", func() {
			Expect(cntg.AllByGUIDs()).To(HaveLen(3))
			Expect(cntg.AllByGUIDs()).To(HaveKeyWithValue("guid1", owg[0]))
			Expect(cntg.AllByGUIDs()).To(HaveKeyWithValue("guid2", owg[1]))
			Expect(cntg.AllByGUIDs()).To(HaveKeyWithValue("guid3", owg[2]))
		})
		It("does not contain object with name 'name1'", func() {
			Expect(cntg.GetByName("name1")).To(BeNil())
		})
		It("returns an empyt list of names", func() {
			Expect(cntg.GetNames()).To(BeEmpty())
		})
		It("is not empty", func() {
			Expect(cntg.IsEmpty()).To(BeFalse())
		})
	})
	Describe("after storing a single ObjectWithName", func() {
		var own *testOWN

		BeforeEach(func() {
			own = &testOWN{"name1"}
			cntn.Store(own)
		})
		It("does not contain object with GUID 'guid1'", func() {
			Expect(cntn.GetByGUID("guid1")).To(BeNil())
		})
		It("returns an empty list of GUIDs", func() {
			Expect(cntn.GetGUIDs()).To(BeEmpty())
		})
		It("can be iterated by GUID", func() {
			Expect(cntn.AllByGUIDs()).To(BeEmpty())
		})
		It("can return it by name", func() {
			Expect(cntn.GetByName("name1")).To(Equal([]*testOWN{
				own,
			}))
			Expect(cntn.GetByName("unknown")).To(BeNil())
		})
		It("can list the single name", func() {
			Expect(cntn.GetNames()).To(Equal([]string{"name1"}))
		})
		It("can be iterated by names", func() {
			Expect(cntn.AllByNames()).To(HaveLen(1))

			Expect(cntn.AllByNames()).To(HaveKeyWithValue("name1", []*testOWN{
				own,
			}))
		})
		It("is not empty", func() {
			Expect(cntn.IsEmpty()).To(BeFalse())
		})
	})
	Describe("after storing multiple mixed Objects", func() {
		var own []mkcontainer.Item

		BeforeEach(func() {
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

		It("can return the elements by GUID", func() {
			Expect(cnt.GetByGUID("guid1")).To(Equal(own[3]))
			Expect(cnt.GetByGUID("guid2")).To(Equal(own[4]))
			Expect(cnt.GetByGUID("guid3")).To(Equal(own[5]))
			Expect(cnt.GetByGUID("unknown")).To(BeNil())
		})
		It("can list the GUIDs", func() {
			Expect(cnt.GetGUIDs()).To(Equal([]string{
				"guid1",
				"guid2",
				"guid3",
			}))
		})
		It("can be iterated by GUID", func() {
			Expect(cnt.AllByGUIDs()).To(HaveLen(3))
			Expect(cnt.AllByGUIDs()).To(HaveKeyWithValue("guid1", own[3]))
			Expect(cnt.AllByGUIDs()).To(HaveKeyWithValue("guid2", own[4]))
			Expect(cnt.AllByGUIDs()).To(HaveKeyWithValue("guid3", own[5]))
		})
		It("can return it by name", func() {
			Expect(cnt.GetByName("name1")).To(Equal([]mkcontainer.Item{
				own[0],
				own[2],
			}))
			Expect(cnt.GetByName("name2")).To(Equal([]mkcontainer.Item{
				own[1],
			}))
			Expect(cnt.GetByName("unknown")).To(BeNil())
		})
		It("can list the names", func() {
			Expect(cnt.GetNames()).To(Equal([]string{
				"name1", "name2",
			}))
		})
		It("can be iterated by names", func() {
			Expect(cnt.AllByNames()).To(HaveLen(2))
			Expect(cnt.AllByNames()).To(HaveKeyWithValue("name1", []mkcontainer.Item{
				own[0],
				own[2],
			}))
			Expect(cnt.AllByNames()).To(HaveKeyWithValue("name2", []mkcontainer.Item{
				own[1],
			}))
		})
		It("is not empty", func() {
			Expect(cnt.IsEmpty()).To(BeFalse())
		})
	})
})
