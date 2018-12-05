package Emulator

import (
  "path/filepath"
  
  
  "github.com/cuu/gogame/color"
	"github.com/cuu/gogame/surface"
  "github.com/cuu/LauncherGoDev/sysgo/UI"

)

type RomListPage struct {
  UI.Page
  Icons  map[string]UI.IconItemInterface
  ListFont *ttf.Font
  MyStack *EmuStack
  EmulatorConfig *ActionConfig
  
  RomSoConfirmDownloadPage *RomSoConfirmPage
  
  MyList []*EmulatorListItem
  BGwidth int
  BGheight 70
  
  Leader *MyEmulator
  
}

func NewRomListPage() *RomListPage {
  p := &RomListPage{}
  p.PageIconMargin = 20
	p.SelectedIconTopOffset = 20
	p.EasingDur = 10

	p.Align = ALIGN["SLeft"]
	
	p.FootMsg = [5]string{ "Nav","Scan","Del","AddFav","Run" }
  
  p.Icons=make(map[string]UI.IconItemInterface)
  p.ListFont =  UI.Fonts["notosanscjk15"]
  
  p.MyStack = NewEmuStack()
  
  p.BGwidth = 56
  p.BGheight = 70
  
  return p
}


func (self *RomListPage) GeneratePathList(path string) []map[string]string {
  if UI.IsDirectory(path) == false {
    return nil
  }
  dirmap := make(map[string]string)
  var ret []map[string]string
  
  file_paths,err := filepath.Glob(path+"/*")//sorted
  if err != nil {
    fmt.Println(err)
    return false
  }
  
  for i,v := range file_paths {
    if UI.IsDirectory(v) && self.EmulatorConfig.FILETYPE == "dir" { // like DOSBOX
      gameshell_bat := self.EmulatorConfig.EXT[0]
      if UI.GetGid(v) == FavGID { // skip fav roms
        continue
      }
      
      if UI.FileExists( filepath.Join(v,gameshell_bat))  == true {
        dirmap["gamedir"] = v
        ret = append(ret,dirmap)
      }
    }
    
    if UI.IsFile(v) && self.EmulatorConfig.FILETYPE == "file" {
      if UI.GetGid(v) == FavGID {
        continue
      }
      
      bname := filepath.Base(v)
      if len(bname) > 1 {
        is_excluded := false
        for _,exclude_pattern := range self.EmulatorConfig.EXCLUDE {
          if matched, err := regexp.MatchString(exclude_pattern,bname); err == nil {
            if matched == true {
              is_excluded = true
              break
            }
          }
        }
        
        if is_excluded == false {
          pieces := strings.Split(bname,".")
          if len(pieces) > 1 {
            pieces_ext := strings.ToLower( pieces[len(pieces)-1])
            for _,v := range self.EmulatorConfig.EXT {
              if pieces_ext == v {
                dirmap["file"] = v
                ret = append(ret,dirmap)
                break
              }
            }
          }
        }
      }
    }
  }
  
  return ret
  
}

func (self *RomListPage) SyncList( path string ) {
  
  alist := self.GeneratePathList(path) 

  if alist == nil {
    fmt.Println("listfiles return false")
    return
  }
  
  self.MyList = nil 
  
  start_x := 0 
  start_y := 0
  
  hasparent := 0 
  
  if self.MyStack.Length() > 0 {
    hasparent = 1
    
    li := NewEmulatorListItem()
    li.Parent = self
    li.PosX   = start_x
    li.PosY   = start_y
    li.Width  = UI.Width
    li.Fonts["normal"] = self.ListFont
    li.MyType = UI.ICON_TYPES["DIR"]
    li.Init("[..]")
    self.MyList = append(self.MyList,li)
  }
  
  for i,v := range alist {
    li := NewEmulatorListItem()
    li.Parent = self
    li.PosX   = start_x
    li.PosY   = start_y + (i+hasparent)*li.Height
    li.Fonts["normal"] = self.ListFont
    li.MyType = UI.ICON_TYPES["FILE"]
    
    init_val = "NoName"
    
    if val, ok := v["directory"]; ok {
      li.MyType = UI.ICON_TYPES["DIR"]
      init_val = val
    }
    
    if val, ok := v["file"]; ok {
      init_val = val
    }
    
    if val, ok := v["gamedir"]; ok {
      init_val = val
    }
    
    li.Init(init_val)
    
    self.MyList = append(self.MyList,li)
  }
}

func (self *RomListPage) Init() {
  self.PosX = self.Index *self.Screen.Width
  self.Width = self.Screen.Width
  self.Height = self.Screen.Height
  
  sefl.CanvasHWND = self.Screen.CanvasHWND
  
  ps := UI.NewInfoPageSelector()
  ps.Width  = UI.Width - 12
  ps.PosX = 2
  ps.Parent = self
  
  self.Ps = ps
  self.PsIndex = 0
  
  self.SyncList( self.EmulatorConfig.ROM )
  
  err := os.MkdirAll( self.EmulatorConfig.ROM+"/.Trash", 0700)
  if err != nil {
    panic(err)
  }
  
  err = os.MkdirAll( self.EmulatorConfig.ROM+"/.Fav", 0700)
  if err != nil {
    panic(err)
  }
  
  self.MyStack.EmulatorConfig = self.EmulatorConfig
  
  icon_for_list := UI.NewMultiIconItem()
  icon_for_list.ImgSurf = UI.MyIconPool.GetImgSurf("sys")
  icon_for_list.MyType = UI.ICON_TYPES["STAT"]
  icon_for_list.Parent = self
  
  icon_for_list.Adjust(0,0,18,18,0)
        
  self.Icons["sys"] = icon_for_list  
  
  bgpng := UI.NewIconItem()
  bgpng.ImgSurf = UI.MyIconPool.GetImgSurf("empty")
  bgpng.MyType = UI.ICON_TYPES["STAT"]
  bgpng.Parent = self
  bgpng.AddLabel("Please upload data over Wi-Fi",UI.Fonts["varela22"])
  bgpng.SetLableColor(&color.Color{204,204,204,255}  )
  bgpng.Adjust(0,0,self.BGwidth,self.BGheight,0)

  self.Icons["bg"] = bgpng
  
  self._Scroller = UI.NewListScroller()
  self._Scroller.Parent = self
  self._Scroller.PosX = self.Width - 10
  self._Scroller.PosY = 2
  self._Scroller.Init()
  
  rom_so_confirm_page := NewRomSoConfirmPage()
  rom_so_confirm_page.Screen = self.Screen
  rom_so_confirm_page.Name = "Download Confirm"
  rom_so_confirm_page.Parent = self
  rom_so_confirm_page.Init()

  self.RomSoConfirmDownloadPage = rom_so_confirm_page 
}


func (self *RomListPage) ScrollUp() {
  if len(self.MyList) == 0 {
    return
  }
  
  self.PsIndex -=1
  
  if self.PsIndex < 0 {
    self.PsIndex = 0
  }
  
  cur_li := self.MyList[self.PsIndex]
  
  if cur_li.PosY < 0 {
    for i,_ := range self.MyList{
      self.MyList[i].PosY += self.MyList[i].Height
    }
    
    self.Scrolled +=1
  }
}


func (self *RomListPage) ScrollDown(){
  if len(self.MyList) == 0 {
    return
  }
  self.PsIndex +=1
  
  if self.PsIndex >= len(self.MyList) {
    self.PsIndex = len(self.MyList) - 1
  }
  
  cur_li := self.MyList[self.PsIndex]
  
  if cur_li.PosY + cur_li.Height > self.Height { 
    for i,_ := range self.MyList{
      self.MyList[i].PosY -= self.MyList[i].Height
    }
    self.Scrolled -=1    
  }

}

func (self *RomListPage) SyncScroll() {

  if self.Scrolled == 0 {
    return
  }
  
  if self.PsIndex < len(self.MyList) {
    cur_li := self.MyList[self.PsIndex]
    if self.Scrolled > 0 {
      if cur_li.PosY < 0 {
        for i,_ := range self.MyList{
          self.MyList[i].PosY += self.Scrolled*self.MyList[i].Height
        }
      }
    } if self.Scrolled < 0 {
      if cur_li.PosY + cur_li.Height > self.Height{
        for i,_ := range self.MyList{
          self.MyList[i].PosY += self.Scrolled*self.MyList[i].Height
        }
      }
    }
  
  }
}


func (self *RomListPage) Click() {

  if len(self.MyList) == 0 {
    return
  }
  
  
  if self.PsIndex > len(self.MyList) - 1 {
    return
  }
  
  
  cur_li := self.MyList[self.PsIndex]
  
  if cur_li.MyType == UI.ICON_TYPES["DIR"] {
    if cur_li.Path = "[..]"{
      self.MyStack.Pop()
      self.SyncList(self.MyStack.Last())
      self.PsIndex = 0
    }else{
      self.MyStack.Push(self.MyList[self.PsIndex].Path)
      self.SyncList(self.MyStack.Last())
      self.PsIndex = 0
    }
  }
  
  if cur_li.MyType == UI.ICON_TYPES["FILE"] {
    self.Screen.MsgBox.SetText("Launching")
    self.Screen.MsgBox.Draw()
    self.Screen.SwapAndShow()
    
    path := ""
    if self.EmulatorConfig.FILETYPE == "dir" {
      path = filepath.Join(cur_li.Path,self.EmulatorConfig.EXT[0])
    }else{
      path  = cur_li.Path
    }
    
    fmt.Println("Run ",path)
    
    escaped_path := UI.CmdClean(path)
    
    if self.EmulatorConfig.FILETYPE == "dir" {
      escaped_path = UI.CmdClean(path)
    }
    
    custom_config := ""
    
    if self.EmulatorConfig.RETRO_CONFIG != "" && len(self.EmulatorConfig.RETRO_CONFIG) 5 {
      custom_config = " -c " + self.EmulatorConfig.RETRO_CONFIG
    }
    
    partsofpath := []string{self.EmulatorConfig.LAUNCHER,self.EmulatorConfig.ROM_SO,custom_config,escaped_path}
    
    cmdpath := strings.Join( partsofpath," ")
    
    if self.EmulatorConfig.ROM_SO =="" { //empty means No needs for rom so 
      event.POST(UI.RUNEVT,cmdpath)
    }else{
      
      if UI.FileExists(self.EmulatorConfig.ROM_SO) == true {
        event.POST(UI.RUNEVT,cmdpath)
      } else {
        self.Screen.PushCurPage()
        self.Screen.SetCurPage( self.RomSoConfirmDownloadPage)
        self.Screen.Draw()
        self.Screen.SwapAndShow()
      }
    }
    
    return
    
  }
  
  self.Screen.Draw()
  self.Screen.SwapAndShow() 
}

func (self *RomListPage) ReScan() {
  if self.MyStack.Length() == 0 {
    self.SyncList(self.EmulatorConfig.ROM)
  }else{
    self.SyncList(self.MyStack.Last())
  }
  
  
  idx := self.PsIndex
  
  if idx > len(self.MyList) - 1 {
    idx = len(self.MyList)
    if idx > 0 {
      idx -= 1
    }else if idx == 0 {
      //nothing in MyList
    }
  }
  
  self.PsIndex = idx //sync PsIndex
  
  self.SyncScroll()
}


func (self *RomListPage) OnReturnBackCb() {
  self.ReScan()
  self.Screen.Draw()
  self.Screen.SwapAndShow()
}


func (self *RomListPage) KeyDown(ev *event.Event) {

  if ev.Data["Key"] == UI.CurKeys["Menu"]{
    self.ReturnToUpLevelPage()
    self.Screen.Draw()
    self.Screen.SwapAndShow()
  }
  
  if ev.Data["Key"] == UI.CurKeys["Right"]{
    self.Screen.PushCurPage()
    self.Screen.SetCurPage(self.Leader.FavPage)
    self.Screen.Draw()
    self.Screen.SwapAndShow()
        
  if ev.Data["Key"] == UI.CurKeys["Up"]{
    self.ScrollUp()
    self.Screen.Draw()
    self.Screen.SwapAndShow()
  }
  
  if ev.Data["Key"] == UI.CurKeys["Down"] {
    self.ScrollDown()
    self.Screen.Draw()
    self.Screen.SwapAndShow()
  }
  
  if ev.Data["Key"] == UI.CurKeys["Enter"] {
    self.Click()
  }
  
  if ev.Data["Key"] == UI.CurKeys["A"] {
    if len(self.MyList) == 0{
      return
    }
    
    cur_li := self.MyList[self.PsIndex]
    
    if cur_li.IsFile() {
      cmd := exec.Command("chgrp", FavGname, UI.CmdClean(cur_li.Path))
      err := cmd.Run()
      if err != nil {
        fmt.Println(err)
      }
      
      self.Screen.MsgBox.SetText("Add to favourite list")
      self.Screen.MsgBox.Draw()
      self.Screen.SwapAndShow()
      time.BlockDelay(600)
      self.ReScan()
      self.Screen.Draw()
      self.Screen.SwapAndShow()
    
    }
  }
  
  if ev.Data["Key"] == UI.CurKeys["X"] { //Scan current
    self.ReScan()
    self.Screen.Draw()
    self.Screen.SwapAndShow()        
  }
  
  if ev.Data["Key"] == UI.CurKeys["Y"] {// del
    if len(self.MyList) == 0 {
      return
    }
    
    cur_li := self.MyList[self.PsIndex] 
    if cur_li.IsFile() {
      self.Leader.DeleteConfirmPage.SetFileName(cur_li.Path)
      self.Leader.DeleteConfirmPage.SetTrashDir(filepath.Join(self.EmulatorConfig.ROM,"/.Trash") )
      
      self.Screen.PushCurPage()
      self.Screen.SetCurPage(self.Leader.DeleteConfirmPage)
      self.Screen.Draw()
      self.Screen.SwapAndShow()
      
    }
  }
}

func (self *RomListPage) Draw() {
  self.ClearCanvas()
  
  if len(self.MyList) == 0 {
    self.Icons["bg"].NewCoord(self.Width/2,self.Height/2)
    self.Icons["bg"].Draw()
  }else{
    
    if len(self.MyList) * HierListItemDefaultHeight > self.Height {
      self.Ps.Width  = self.Width - 10
      self.Ps.Draw()
      
      
      for i,v := range self.MyList {
        if v.PosY > self.Height + self.Height/2 {
          break
        }
        
        if v.PosY < 0 {
          continue
        }
        
        v.Draw()
      }
      
      self.Scroller.UpdateSize( len(self.MyList)*HierListItemDefaultHeight, self.PsIndex*HierListItemDefaultHeight)
      self.Scroller.Draw()
      
      
      
    }else {
      self.Ps.Width = self.Width
      self.Ps.Draw()
      for _,v := range self.MyList {
        v.Draw()
      }
    }
  }
}

